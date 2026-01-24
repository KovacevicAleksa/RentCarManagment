package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
	"github.com/KovacevicAleksa/rentcar/backend/internal/db"
	"github.com/KovacevicAleksa/rentcar/backend/internal/mqtt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Database setup
	dbConn := db.NewPostgresConnection()
	if err := dbConn.AutoMigrate(&auth.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Auth setup
	authRepo := auth.NewAuthRepository(dbConn)
	authService := auth.NewAuthService(authRepo)
	auth.RegisterRoutes(r, authService)

	// MQTT setup
	mqttClient := mqtt.NewClient()
	mqttService := mqtt.NewService(mqttClient)

	topics := map[string]mqtt.MessageHandler{
		"car/telemetry": mqtt.CarTelemetryHandler,
	}

	if err := mqttService.Start(topics); err != nil {
		log.Printf("MQTT service failed to start: %v", err)
	}

	log.Println("Server starting on :8010")
	r.Run(":8010")
}