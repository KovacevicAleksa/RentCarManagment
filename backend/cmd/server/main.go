package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
	"github.com/KovacevicAleksa/rentcar/backend/internal/db"
	"github.com/KovacevicAleksa/rentcar/backend/internal/history"
	"github.com/KovacevicAleksa/rentcar/backend/internal/mqtt"
	"github.com/KovacevicAleksa/rentcar/backend/internal/websocket"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	allowedOrigins := []string{frontendURL}
	if adminURL := os.Getenv("ADMIN_URL"); adminURL != "" {
		allowedOrigins = append(allowedOrigins, adminURL)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	postgresConn := db.NewPostgresConnection()
	timescaleConn := db.NewTimescaleConnection()

	if err := postgresConn.AutoMigrate(&auth.User{}); err != nil {
		log.Fatal("Failed to migrate Postgres:", err)
	}

	if err := db.EnableTimescaleDB(timescaleConn); err != nil {
		log.Printf("TimescaleDB error: %v", err)
	}

	if err := timescaleConn.AutoMigrate(&history.CarHistory{}); err != nil {
		log.Fatal("Failed to migrate TimescaleDB:", err)
	}

	if err := db.ConvertToHypertable(timescaleConn, "car_histories", "timestamp"); err != nil {
		log.Printf("Hypertable error: %v", err)
	} else {
		db.CreateCompressionPolicy(timescaleConn, "car_histories", "7 days")
		db.CreateRetentionPolicy(timescaleConn, "car_histories", "90 days")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	tokenService := auth.NewTokenService(jwtSecret, 72*time.Hour)

	authRepo := auth.NewAuthRepository(postgresConn)
	authService := auth.NewAuthService(authRepo, tokenService)

	if err := authService.EnsureAdmin(os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")); err != nil {
		log.Printf("Failed to seed admin user: %v", err)
	}

	auth.RegisterRoutes(r, authService, tokenService)

	historyRepo := history.NewHistoryRepository(timescaleConn)
	historyService := history.NewHistoryService(historyRepo)
	history.RegisterRoutes(r, historyService, tokenService)

	wsHub := websocket.NewHub()
	go wsHub.Run()

	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(wsHub, c)
	})

	mqtt.SetBroadcaster(wsHub)
	mqtt.SetHistorySaver(historyService)

	mqttClient := mqtt.NewClient()
	mqttService := mqtt.NewService(mqttClient)

	topics := map[string]mqtt.MessageHandler{
		"car/+/telemetry": mqtt.CarTelemetryHandler,
	}

	if err := mqttService.Start(topics); err != nil {
		log.Printf("MQTT failed: %v", err)
	}

	log.Println("Server starting on :8010")
	r.Run(":8010")
}