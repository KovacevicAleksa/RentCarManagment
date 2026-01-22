package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/KovacevicAleksa/rentcar/backend/internal/db"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Hello from golang")
    db := db.NewPostgresConnection()
        defer db.Close()

}
