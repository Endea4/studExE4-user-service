package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Endea4/studExE4-shared/auth"
	"github.com/Endea4/studExE4-user-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func registerUserRoutes(r *gin.Engine, usersColl *mongo.Collection) {
	r.GET("/users", func(c *gin.Context) {
		users, err := userRepo.GetAllUsers(context.Background())
		if err != nil {
			logErr("get all users failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	r.GET("/users/:phone", func(c *gin.Context) {
		phone := c.Param("phone")
		user, err := userRepo.GetUserByPhone(c.Request.Context(), phone)
		if err != nil {
			if objID, oidErr := primitive.ObjectIDFromHex(phone); oidErr == nil {
				user, err = userRepo.GetUserByID(c.Request.Context(), objID.Hex())
			}
			if err != nil {
				logErr("get user failed: param=%s err=%v", phone, err)
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
		}
		c.JSON(http.StatusOK, user)
	})

	r.PUT("/users/:phone", func(c *gin.Context) {
		phone := c.Param("phone")
		var req struct {
			Name        *string `json:"name"`
			DisplayName *string `json:"display_name"`
			Email       *string `json:"email"`
			Gender      *string `json:"gender"`
			VehicleType *string `json:"vehicle_type"`
			PlateNumber *string `json:"plate_number"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			logErr("update user bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{"updated_at": time.Now()}
		if req.Name != nil {
			update["name"] = *req.Name
		}
		if req.DisplayName != nil {
			update["display_name"] = *req.DisplayName
		}
		if req.Email != nil {
			update["email"] = *req.Email
		}
		if req.Gender != nil {
			update["gender"] = *req.Gender
		}
		if req.VehicleType != nil {
			update["vehicle_info.vehicle_type"] = *req.VehicleType
		}
		if req.PlateNumber != nil {
			update["vehicle_info.license_plate"] = *req.PlateNumber
		}

		if err := userRepo.UpdateUser(c.Request.Context(), phone, update); err != nil {
			logErr("update user failed: phone=%s err=%v", phone, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		user, err := userRepo.GetUserByPhone(c.Request.Context(), phone)
		if err != nil {
			logErr("get user after update failed: phone=%s err=%v", phone, err)
		}
		c.JSON(http.StatusOK, user)
	})

	r.PUT("/users/:phone/role", func(c *gin.Context) {
		phone := c.Param("phone")
		var req struct {
			Role string `json:"role" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := userRepo.GetUserByPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		for _, r := range user.Roles {
			if r == req.Role {
				c.JSON(http.StatusConflict, gin.H{"error": "user already has this role"})
				return
			}
		}

		if req.Role == models.RoleDriver {
			if err := userRepo.PromoteDriver(c.Request.Context(), phone, "", ""); err != nil {
				logErr("promote driver failed: phone=%s err=%v", phone, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to promote driver"})
				return
			}
		} else {
			if err := userRepo.UpdateUser(c.Request.Context(), phone, bson.M{
				"$addToSet": bson.M{"roles": req.Role},
				"updated_at": time.Now(),
			}); err != nil {
				logErr("role update failed: phone=%s err=%v", phone, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
				return
			}
		}

		updatedUser, _ := userRepo.GetUserByPhone(c.Request.Context(), phone)
		token, _ := auth.GenerateJWT(user.ID.Hex(), phone, strings.Join(updatedUser.Roles, ","))
		c.JSON(http.StatusOK, gin.H{
			"user":  updatedUser,
			"token": token,
		})
	})

	r.PUT("/users/:phone/password", func(c *gin.Context) {
		phone := c.Param("phone")
		var req struct {
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hashed, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		if err := userRepo.UpdateUser(c.Request.Context(), phone, bson.M{"password": hashed, "updated_at": time.Now()}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "password updated"})
	})

	r.POST("/users/:phone/saved-locations", func(c *gin.Context) {
		phone := c.Param("phone")
		var loc models.SavedLocation
		if err := c.ShouldBindJSON(&loc); err != nil {
			logErr("saved-location bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		filter := bson.M{"phone": phone}
		update := bson.M{
			"$push": bson.M{"saved_locations": loc},
			"$set":  bson.M{"updated_at": time.Now()},
		}
		res, err := usersColl.UpdateOne(c.Request.Context(), filter, update)
		if err != nil {
			logErr("add saved location failed: phone=%s err=%v", phone, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add saved location"})
			return
		}
		if res.MatchedCount == 0 {
			logErr("add saved location: user not found phone=%s", phone)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "saved location added"})
	})

	r.GET("/users/:phone/saved-locations", func(c *gin.Context) {
		phone := c.Param("phone")
		filter := bson.M{"phone": phone}
		proj := options.FindOne().SetProjection(bson.M{"saved_locations": 1, "_id": 0})
		var result struct {
			SavedLocations []models.SavedLocation `bson:"saved_locations" json:"saved_locations"`
		}
		err := usersColl.FindOne(c.Request.Context(), filter, proj).Decode(&result)
		if err != nil {
			logErr("get saved locations failed: phone=%s err=%v", phone, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, result.SavedLocations)
	})

	r.DELETE("/users/:phone/saved-locations/:name", func(c *gin.Context) {
		phone := c.Param("phone")
		name := c.Param("name")
		filter := bson.M{"phone": phone}
		update := bson.M{
			"$pull": bson.M{"saved_locations": bson.M{"name": name}},
			"$set":  bson.M{"updated_at": time.Now()},
		}
		res, err := usersColl.UpdateOne(c.Request.Context(), filter, update)
		if err != nil {
			logErr("remove saved location failed: phone=%s err=%v", phone, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove saved location"})
			return
		}
		if res.MatchedCount == 0 {
			logErr("remove saved location: user not found phone=%s", phone)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "saved location removed"})
	})
}
