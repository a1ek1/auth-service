package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var rdb *redis.Client
var jwtSecret = []byte("secret_key_for_jwt")

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "redis:6379", // Адрес Redis
		DB:   0,            // Используем 0-ю базу данных
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь в Redis
	ctx := context.Background()
	_, err = rdb.Get(ctx, user.Login).Result()
	if err == nil {
		log.Println("User already exists:", user.Login)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	} else if err != redis.Nil {
		log.Println("Error checking user in Redis:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Хешируем пароль
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Сохраняем в Redis
	err = rdb.Set(ctx, user.Login, hashedPassword, 0).Err()
	if err != nil {
		log.Println("Error saving to Redis:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Response{Message: "User registered successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	storedPassword, err := rdb.Get(ctx, loginReq.Login).Result()
	if err == redis.Nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !utils.CheckPasswordHash(loginReq.Password, storedPassword) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": loginReq.Login,
		"exp":   time.Now().Add(time.Hour * 1).Unix(), // срок действия 1 час
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = rdb.Set(ctx, fmt.Sprintf("token:%s", loginReq.Login), tokenString, time.Hour).Err()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(models.AuthResponse{Token: tokenString})
}

// Страница успешного входа
func SuccessHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token not provided", http.StatusUnauthorized)
		return
	}

	// Парсим JWT токен
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !parsedToken.Valid {
		log.Printf("Invalid token: %v", err) // Логирование ошибки токена
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Извлекаем login из токена
	login, ok := claims["login"].(string)
	if !ok {
		log.Println("Invalid token structure: 'login' not found or not a string") // Логирование ошибки структуры токена
		http.Error(w, "Invalid token structure", http.StatusUnauthorized)
		return
	}

	// Проверяем наличие токена в Redis
	ctx := context.Background()
	storedToken, err := rdb.Get(ctx, fmt.Sprintf("token:%s", login)).Result()

	if err == redis.Nil {
		log.Printf("Token not found in Redis for login: %s", login) // Логирование случая, когда токен не найден в Redis
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	if err != nil {
		log.Printf("Error retrieving token from Redis: %v", err) // Логирование ошибки при доступе к Redis
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Если токен совпадает
	if storedToken == token {
		log.Printf("Token validated successfully for user: %s", login) // Логирование успешной валидации токена
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Response{Message: "Successfully logged in"})
	} else {
		log.Printf("Token mismatch for user: %s", login) // Логирование несоответствия токенов
		http.Error(w, "Invalid token", http.StatusUnauthorized)
	}
}
