package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/utils"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var rdb *redis.Client
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
		DB:   0,
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		utils.SendJSONResponse(w, http.StatusBadRequest, models.Response{Message: "Invalid input"})
		return
	}

	// Проверка на пустой логин и пароль
	if strings.TrimSpace(user.Login) == "" || strings.TrimSpace(user.Password) == "" {
		log.Println("Empty login or password provided")
		utils.SendJSONResponse(w, http.StatusBadRequest, models.Response{Message: "Login and password cannot be empty"})
		return
	}

	log.Printf("Decoded user: %+v\n", user)

	// Проверяем, существует ли пользователь с таким логином в PostgreSQL
	ctx := context.Background()
	var exists bool
	err = utils.DB.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE login=$1)", user.Login).Scan(&exists)
	if err != nil {
		log.Println("Error querying PostgreSQL:", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	if exists {
		log.Println("User already exists:", user.Login)
		utils.SendJSONResponse(w, http.StatusConflict, models.Response{Message: "User already exists"})
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	// Сохраняем пользователя в PostgreSQL
	_, err = utils.DB.Exec(ctx, "INSERT INTO users (login, password) VALUES ($1, $2)", user.Login, hashedPassword)
	if err != nil {
		log.Println("Error inserting user into PostgreSQL:", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, models.Response{Message: "User registered successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, models.Response{Message: "Invalid input"})
		return
	}

	ctx := context.Background()
	var storedPassword string
	err = utils.DB.QueryRow(ctx, "SELECT password FROM users WHERE login=$1", loginReq.Login).Scan(&storedPassword)
	if err == pgx.ErrNoRows {
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "User not found"})
		return
	}
	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	if !utils.CheckPasswordHash(loginReq.Password, storedPassword) {
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": loginReq.Login,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	err = rdb.Set(ctx, fmt.Sprintf("token:%s", loginReq.Login), tokenString, time.Hour).Err()
	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	utils.SendJSONResponse(w, http.StatusOK, models.Response{Message: "Login successful"})
}

// Страница успешного входа
func SuccessHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "Token not provided"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")

	// Парсим JWT токен
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !parsedToken.Valid {
		log.Printf("Invalid token: %v", err)
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "Invalid token"})
		return
	}

	login, ok := claims["login"].(string)
	if !ok {
		log.Println("Invalid token structure: 'login' not found or not a string")
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "Invalid token structure"})
		return
	}

	// Проверяем наличие токена в Redis
	ctx := context.Background()
	storedToken, err := rdb.Get(ctx, fmt.Sprintf("token:%s", login)).Result()

	if err == redis.Nil {
		log.Printf("Token not found in Redis for login: %s", login)
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "Invalid token"})
		return
	}

	if err != nil {
		log.Printf("Error retrieving token from Redis: %v", err)
		utils.SendJSONResponse(w, http.StatusInternalServerError, models.Response{Message: "Internal server error"})
		return
	}

	// Если токен совпадает
	if storedToken == token {
		log.Printf("Token validated successfully for user: %s", login)
		utils.SendJSONResponse(w, http.StatusOK, models.Response{Message: "Successfully logged in"})
	} else {
		log.Printf("Token mismatch for user: %s", login)
		utils.SendJSONResponse(w, http.StatusUnauthorized, models.Response{Message: "Invalid token"})
	}
}
