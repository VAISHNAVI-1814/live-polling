package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"live-polling-backend/config"
	"live-polling-backend/controllers"
	"live-polling-backend/middleware"
	"live-polling-backend/repository"
	"live-polling-backend/services"
	"live-polling-backend/websocket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Printf("🚀 Starting Live Polling Server on port %s...\n", cfg.Port)

	// Connect to MongoDB
	mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer mongoCancel()

	isMongoAvailable := false
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(mongoCtx, clientOptions)
	if err == nil {
		if pingErr := mongoClient.Ping(mongoCtx, nil); pingErr != nil {
			log.Printf("⚠️ MongoDB Ping failed at %s (%v). Running with high-performance In-Memory Repository fallback.", cfg.MongoURI, pingErr)
		} else {
			log.Printf("✅ Connected to MongoDB successfully at %s", cfg.MongoURI)
			isMongoAvailable = true
		}
	} else {
		log.Printf("⚠️ MongoDB connection error: %v. Running with high-performance In-Memory Repository fallback.", err)
	}

	db := mongoClient.Database(cfg.MongoDBName)

	// Repositories
	userRepo := repository.NewUserRepository(db, isMongoAvailable)
	pollRepo := repository.NewPollRepository(db, isMongoAvailable)

	// Services
	redisSvc := services.NewRedisService(cfg.RedisURL, cfg.RedisPassword)
	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret)
	pollSvc := services.NewPollService(pollRepo, redisSvc)

	// WebSocket Hub
	wsHub := websocket.NewHub(redisSvc)
	go wsHub.Run()

	// Controllers
	authCtrl := controllers.NewAuthController(authSvc)
	pollCtrl := controllers.NewPollController(pollSvc, wsHub)

	// Gin Router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg.ClientOrigin))

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"mongo":     isMongoAvailable,
			"redis":     redisSvc.IsAvailable(),
		})
	})

	// WebSocket route for real-time live poll updates
	router.GET("/ws/polls/:id", pollCtrl.ServeWS)

	api := router.Group("/api")
	{
		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authCtrl.Signup)
			auth.POST("/login", authCtrl.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authCtrl.Me)
		}

		// Public poll endpoints
		polls := api.Group("/polls")
		{
			polls.GET("/:id", pollCtrl.GetPoll)
			polls.GET("/share/:shareCode", pollCtrl.GetPollByShareCode)
			polls.POST("/:id/vote", pollCtrl.Vote)
			polls.GET("/:id/results", pollCtrl.GetResults)
			polls.GET("/:id/voted", pollCtrl.CheckVoted)

			// Authenticated poll management
			protected := polls.Group("")
			protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				protected.POST("", pollCtrl.CreatePoll)
				protected.GET("/my-polls", pollCtrl.GetMyPolls)
				protected.PATCH("/:id/status", pollCtrl.UpdateStatus)
				protected.DELETE("/:id", pollCtrl.DeletePoll)
			}
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("✅ Live Polling Server listening on http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
