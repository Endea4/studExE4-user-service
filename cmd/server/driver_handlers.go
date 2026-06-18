package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Endea4/studExE4-shared/auth"
	"github.com/Endea4/studExE4-user-service/internal/models"
	"github.com/Endea4/studExE4-user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func registerDriverRoutes(r *gin.Engine) {
	r.POST("/drivers/add-role", func(c *gin.Context) {
		var req struct {
			Phone       string `json:"phone" binding:"required"`
			VehicleType string `json:"vehicle_type"`
			PlateNumber string `json:"plate_number"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := userRepo.GetUserByPhone(c.Request.Context(), req.Phone)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		for _, r := range user.Roles {
			if r == models.RoleDriver {
				c.JSON(http.StatusConflict, gin.H{"error": "user is already a driver"})
				return
			}
		}

		if err := userRepo.PromoteDriver(c.Request.Context(), req.Phone, req.VehicleType, req.PlateNumber); err != nil {
			logErr("promote driver failed: phone=%s err=%v", req.Phone, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to promote driver"})
			return
		}

		hashedPassword, _ := auth.HashPassword(req.Phone)
		userRepo.UpdateUser(c.Request.Context(), req.Phone, bson.M{"password": hashedPassword, "updated_at": time.Now()})

		updatedUser, _ := userRepo.GetUserByPhone(c.Request.Context(), req.Phone)
		token, _ := auth.GenerateJWT(user.ID.Hex(), req.Phone, strings.Join(updatedUser.Roles, ","))
		c.JSON(http.StatusOK, gin.H{
			"user":  updatedUser,
			"token": token,
		})
	})

	r.GET("/drivers/active", func(c *gin.Context) {
		drivers, err := userRepo.GetActiveDrivers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, drivers)
	})

	r.GET("/drivers/all", func(c *gin.Context) {
		drivers, err := userRepo.GetAllDrivers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, drivers)
	})

	r.GET("/drivers", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone != "" {
			user, err := userRepo.GetUserByPhone(c.Request.Context(), phone)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
				return
			}
			if !isDriver(user) {
				c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
				return
			}
			c.JSON(http.StatusOK, user)
			return
		}

		userID := c.Query("user_id")
		if userID != "" {
			user, err := userRepo.GetUserByID(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
				return
			}
			if !isDriver(user) {
				c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
				return
			}
			c.JSON(http.StatusOK, user)
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "phone or user_id query param required"})
	})

	r.GET("/drivers/:id", func(c *gin.Context) {
		id := c.Param("id")
		user, err := userRepo.GetUserByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	r.PUT("/drivers/status", func(c *gin.Context) {
		var req struct {
			Phone  string `json:"phone"`
			UserID string `json:"user_id"`
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.Status = strings.ToUpper(req.Status)

		phone := req.Phone
		var gender string
		if phone == "" && req.UserID != "" {
			user, err := userRepo.GetUserByID(c.Request.Context(), req.UserID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if len(user.Phone) > 0 {
				phone = user.Phone[0]
			}
			gender = user.Gender
		} else if phone != "" {
			user, _ := userRepo.GetUserByPhone(c.Request.Context(), phone)
			if user != nil {
				gender = user.Gender
			}
		}
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone or user_id required"})
			return
		}

		if err := userRepo.SetDriverStatus(c.Request.Context(), phone, req.Status); err != nil {
			logErr("failed to update driver status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
			return
		}

		isOnline := req.Status == models.DriverStatusReady
		refID := req.UserID
		if refID == "" {
			refID = phone
		}
		publishDriverStatus(refID, phone, gender, isOnline)

		c.JSON(http.StatusOK, gin.H{"message": "status updated"})
	})

	r.PUT("/drivers/online", func(c *gin.Context) {
		var req struct {
			Phone    string `json:"phone"`
			UserID   string `json:"user_id"`
			IsOnline bool   `json:"is_online"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		status := models.DriverStatusNotReady
		if req.IsOnline {
			status = models.DriverStatusReady
		}

		phone := req.Phone
		var gender string
		if phone == "" && req.UserID != "" {
			user, err := userRepo.GetUserByID(c.Request.Context(), req.UserID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			if len(user.Phone) > 0 {
				phone = user.Phone[0]
			}
			gender = user.Gender
		} else if phone != "" {
			user, _ := userRepo.GetUserByPhone(c.Request.Context(), phone)
			if user != nil {
				gender = user.Gender
			}
		}
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone or user_id required"})
			return
		}

		if err := userRepo.SetDriverStatus(c.Request.Context(), phone, status); err != nil {
			logErr("failed to update online status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update online status"})
			return
		}

		refID := req.UserID
		if refID == "" {
			refID = phone
		}
		publishDriverStatus(refID, phone, gender, req.IsOnline)

		c.JSON(http.StatusOK, gin.H{"message": "online status updated"})
	})
}

func registerDriverDebtRoutes(r *gin.Engine, dr *repository.DebtRepository) {
	r.GET("/drivers/debts", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		debts, err := dr.GetByDriverPhone(c.Request.Context(), phone)
		if err != nil {
			logErr("failed to get debts: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get debts"})
			return
		}
		c.JSON(http.StatusOK, debts)
	})

	r.POST("/drivers/debts", func(c *gin.Context) {
		var req struct {
			Phone   string  `json:"phone" binding:"required"`
			Amount  float64 `json:"amount" binding:"required"`
			Note    string  `json:"note"`
			OrderID string  `json:"order_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		debt := &models.Debt{
			DriverPhone: req.Phone,
			Amount:      req.Amount,
			Remaining:   req.Amount,
			Description: req.Note,
			Status:      "unpaid",
		}
		if req.OrderID != "" {
			oid, err := primitive.ObjectIDFromHex(req.OrderID)
			if err != nil {
				logErr("create debt: invalid order_id=%s err=%v", req.OrderID, err)
			} else {
				debt.OrderID = oid
			}
		}
		if err := dr.Create(c.Request.Context(), debt); err != nil {
			logErr("failed to create debt: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create debt"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "debt created"})
	})

	r.PUT("/drivers/debts/:id/pay", func(c *gin.Context) {
		id := c.Param("id")
		if err := dr.MarkPaid(c.Request.Context(), id); err != nil {
			logErr("failed to mark paid: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark paid"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "debt marked as paid"})
	})

	r.PUT("/drivers/debts/:id/dispute", func(c *gin.Context) {
		id := c.Param("id")
		if err := dr.MarkDisputed(c.Request.Context(), id); err != nil {
			logErr("failed to mark disputed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark disputed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "debt marked as disputed"})
	})

	r.GET("/drivers/debts/active", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		hasActive, err := dr.HasActiveDebts(c.Request.Context(), phone)
		if err != nil {
			logErr("failed to check active debts: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check active debts"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"has_active": hasActive})
	})
}

func registerDriverLeaveRoutes(r *gin.Engine, lr *repository.LeaveRequestRepository) {
	r.POST("/drivers/leave-requests", func(c *gin.Context) {
		var req struct {
			Phone  string `json:"phone" binding:"required"`
			Type   string `json:"type" binding:"required"`
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Type != "izin" && req.Type != "cuti" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'izin' or 'cuti'"})
			return
		}
		lr := &models.LeaveRequest{
			DriverPhone: req.Phone,
			Type:        req.Type,
			Reason:      req.Reason,
		}
		if err := leaveRequestRepo.Create(c.Request.Context(), lr); err != nil {
			logErr("failed to create leave request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create leave request"})
			return
		}
		c.JSON(http.StatusCreated, lr)
	})

	r.GET("/drivers/leave-requests", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		results, err := leaveRequestRepo.GetByDriverPhone(c.Request.Context(), phone)
		if err != nil {
			logErr("failed to get leave requests: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get leave requests"})
			return
		}
		c.JSON(http.StatusOK, results)
	})

	r.PUT("/drivers/leave-requests/:id/approve", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			ReviewedBy string `json:"reviewed_by" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := leaveRequestRepo.UpdateStatus(c.Request.Context(), id, "approved", req.ReviewedBy); err != nil {
			logErr("failed to approve leave request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve leave request"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "leave request approved"})
	})

	r.PUT("/drivers/leave-requests/:id/reject", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			ReviewedBy string `json:"reviewed_by" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := leaveRequestRepo.UpdateStatus(c.Request.Context(), id, "rejected", req.ReviewedBy); err != nil {
			logErr("failed to reject leave request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject leave request"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "leave request rejected"})
	})
}

func isDriver(user *models.User) bool {
	for _, r := range user.Roles {
		if r == models.RoleDriver {
			return true
		}
	}
	return false
}
