package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
	"github.com/KovacevicAleksa/rentcar/backend/internal/db"
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

	dbConn := db.NewPostgresConnection()

	if err := dbConn.AutoMigrate(&auth.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	authRepo := auth.NewAuthRepository(dbConn)
	authService := auth.NewAuthService(authRepo)
	auth.RegisterRoutes(r, authService)

	r.Run(":8010")
}
