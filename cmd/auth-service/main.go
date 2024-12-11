package main

import (
	"auth-service/cmd/auth-service/routes"
	"auth-service/internal/utils"
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, loading environment variables from system")
	}

	utils.ConnectToDB()
	defer utils.DB.Close(context.Background())

	utils.RunMigrations()

	routes.SetupRoutes()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
