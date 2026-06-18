package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/Endea4/studExE4-shared/config"
	"github.com/Endea4/studExE4-shared/mongo"
	"github.com/Endea4/studExE4-user-service/internal/consumer"
	"github.com/Endea4/studExE4-user-service/internal/models"
	"github.com/Endea4/studExE4-user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func logErr(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

var (
	userRepo         *repository.UserRepository
	debtRepo         *repository.DebtRepository
	leaveRequestRepo *repository.LeaveRequestRepository
	rdb              *redis.Client
)

func main() {
	config.LoadConfig()
	uri := config.GetEnv("MONGODB_URI", "")
	dbName := config.GetEnv("DB_NAME", "studex-users")
	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "")

	client, db := mongo.ConnectDB(uri, dbName)
	defer mongo.Disconnect(client)

	userRepo = repository.NewUserRepository(db)
	debtRepo = repository.NewDebtRepository(db)
	leaveRequestRepo = repository.NewLeaveRequestRepository(db)

	if err := leaveRequestRepo.EnsureIndexes(context.Background()); err != nil {
		logErr("leave request index creation failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		rdb = nil
	}

	if kafkaBrokers != "" {
		brokers := strings.Split(kafkaBrokers, ",")
		tripConsumer := consumer.NewTripEventConsumer(debtRepo, userRepo)
		tripConsumer.Start(ctx, brokers)
	}

	if rdb != nil {
		tripConsumer := consumer.NewTripEventConsumer(debtRepo, userRepo)
		go tripConsumer.StartRedis(ctx, rdb)
	}

	r := gin.Default()
	usersColl := db.Collection("users")

	registerAuthRoutes(r)
	registerUserRoutes(r, usersColl)
	registerDriverRoutes(r)
	registerDriverDebtRoutes(r, debtRepo)
	registerDriverLeaveRoutes(r, leaveRequestRepo)

	r.GET("/info", func(c *gin.Context) {
		ctx := c.Request.Context()
		total, _ := userRepo.CountByFilter(ctx, bson.M{})
		drivers, _ := userRepo.CountByFilter(ctx, bson.M{"roles": models.RoleDriver})
		ready, _ := userRepo.CountByFilter(ctx, bson.M{"driver_status": models.DriverStatusReady})
		notReady, _ := userRepo.CountByFilter(ctx, bson.M{"driver_status": models.DriverStatusNotReady})
		onJob, _ := userRepo.CountByFilter(ctx, bson.M{"driver_status": models.DriverStatusOnJob})
		suspended, _ := userRepo.CountByFilter(ctx, bson.M{"driver_status": models.DriverStatusSuspended})
		unverified, _ := userRepo.CountByFilter(ctx, bson.M{"is_document_verified": false})
		active, _ := userRepo.CountByFilter(ctx, bson.M{"status": models.StatusActive})
		c.JSON(http.StatusOK, gin.H{
			"service":       "user-service",
			"total_users":   total,
			"drivers":       drivers,
			"ready":         ready,
			"not_ready":     notReady,
			"on_job":        onJob,
			"suspended":     suspended,
			"unverified":    unverified,
			"active":        active,
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "user-service"})
	})

	port := config.GetEnv("PORT", "9086")
	go func() {
		if err := r.Run(":" + port); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	signal.Ignore(syscall.SIGHUP)
	<-quit
	cancel()
}
