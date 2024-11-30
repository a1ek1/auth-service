package main

import (
	"auth-service/cmd/auth-service/routes"
	"auth-service/internal/utils"
	"context"
	"log"
	"net/http"
)

func main() {
	utils.ConnectToDB()
	defer utils.DB.Close(context.Background())

	routes.SetupRoutes()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
