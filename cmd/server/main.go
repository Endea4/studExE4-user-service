package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Endea4/studExE4-user-service/shared/config"
	"github.com/Endea4/studExE4-user-service/shared/mongo"
	"github.com/Endea4/studExE4-user-service/internal/models"
	"github.com/Endea4/studExE4-user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// 1. Load config and connect to MongoDB (specifically the users DB)
	config.LoadConfig()
	uri := config.GetEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := config.GetEnv("DB_NAME", "studexdb") // Using the central DB

	client, db := mongo.ConnectDB(uri, dbName)
	defer mongo.Disconnect(client)

	// 2. Initialize the User Repository
	userRepo := repository.NewUserRepository(db)

	// 3. Set up a simple Gin router to test our CRUD
	r := gin.Default()

	// --- 1. Get all users ---
	r.GET("/users", func(c *gin.Context) {
		users, err := userRepo.GetAllUsers(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	// --- 1.5 Get single user by phone ---
	r.GET("/users/:phone", func(c *gin.Context) {
		phone := c.Param("phone")
		user, err := userRepo.GetUserByPhone(context.Background(), phone)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// --- 2. "Bot registers number" (Create minimal user) ---
	r.POST("/users/register", func(c *gin.Context) {
		var req struct {
			Phone string `json:"phone" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if _, err := userRepo.GetUserByPhone(context.Background(), req.Phone); err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		newUser := &models.User{
			Phone: req.Phone,
			CustomerProfile: &models.Customer{
				ReputationScore: 5.0,
				TotalOrders:     0,
				TotalCancels:    0,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := userRepo.CreateUser(context.Background(), newUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered via Bot", "user": newUser})
	})

	// --- 3. "Bot completes personalization" (Update user) ---
	r.PUT("/users/:phone/personalize", func(c *gin.Context) {
		phone := c.Param("phone")

		var req struct {
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Gender      string `json:"gender"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateData := bson.M{
			"name":         req.Name,
			"display_name": req.DisplayName,
			"gender":       req.Gender,
			"updated_at":   time.Now(),
		}

		if err := userRepo.UpdateUser(context.Background(), phone, updateData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		updatedUser, _ := userRepo.GetUserByPhone(context.Background(), phone)
		c.JSON(http.StatusOK, gin.H{"message": "Personalization complete", "user": updatedUser})
	})

	// --- 4. Delete User (For testing) ---
	r.DELETE("/users/:phone", func(c *gin.Context) {
		phone := c.Param("phone")
		if err := userRepo.DeleteUser(context.Background(), phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	})

	r.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!DOCTYPE html>
<html><head><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
<script>SwaggerUIBundle({url:"/docs/openapi.yaml",dom_id:"#swagger-ui"})</script></body></html>`))
	})
	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")

	port := config.GetEnv("USER_SERVICE_PORT", "9081")
	r.Run(":" + port)
}
