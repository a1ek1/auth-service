package models

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Response struct {
	Message string `json:"message"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
