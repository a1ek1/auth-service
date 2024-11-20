package routes

import (
	"auth-service/cmd/auth-service/handlers"
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/success", handlers.SuccessHandler)
}
