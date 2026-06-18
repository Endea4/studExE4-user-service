package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Endea4/studExE4-shared/auth"
	"github.com/Endea4/studExE4-user-service/internal/models"
)

func registerAuthRoutes(r *gin.Engine) {
	r.POST("/auth/register", func(c *gin.Context) {
		var req struct {
			Phone    string `json:"phone" binding:"required"`
			Password string `json:"password" binding:"required,min=4"`
			Name     string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			logErr("register bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existing, _ := userRepo.GetUserByPhone(c.Request.Context(), req.Phone)
		if existing != nil {
			logErr("register conflict: phone %s already registered", req.Phone)
			c.JSON(http.StatusConflict, gin.H{"error": "phone already registered"})
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			logErr("hash password failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		now := time.Now()
		user := &models.User{
			Phone:       []string{req.Phone},
			Password:    hashedPassword,
			Name:        req.Name,
			DisplayName: req.Name,
			Email:       req.Phone + "@driver.local",
			Roles:       []string{models.RoleUser},
			Status:      models.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := userRepo.CreateUser(c.Request.Context(), user); err != nil {
			logErr("create user failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		phone := ""
		if len(user.Phone) > 0 {
			phone = user.Phone[0]
		}
		token, err := auth.GenerateJWT(user.ID.Hex(), phone, strings.Join(user.Roles, ","))
		if err != nil {
			logErr("generate JWT failed for register: %v", err)
		}
		c.JSON(http.StatusCreated, gin.H{
			"user":  user,
			"token": token,
		})
	})

	r.POST("/auth/login", func(c *gin.Context) {
		var req struct {
			Phone    string `json:"phone" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			logErr("login bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := userRepo.GetUserByPhone(c.Request.Context(), req.Phone)
		if err != nil {
			logErr("login user not found: phone=%s err=%v", req.Phone, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if !auth.CheckPassword(req.Password, user.Password) {
			logErr("login wrong password: phone=%s", req.Phone)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		phone := ""
		if len(user.Phone) > 0 {
			phone = user.Phone[0]
		}
		token, err := auth.GenerateJWT(user.ID.Hex(), phone, strings.Join(user.Roles, ","))
		if err != nil {
			logErr("generate JWT failed for login: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"user":  user,
			"token": token,
		})
	})
}
