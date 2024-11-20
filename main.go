package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

var rdb *redis.Client
var jwtSecret = []byte("secret_key_for_jwt")

// Структуры для входных данных
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Структуры для ответов
type Response struct {
	Message string `json:"message"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// Инициализация Redis
func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Адрес Redis
		DB:   0,                // Используем 0-ю базу данных
	})
}

// Функция для хеширования пароля
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Функция для проверки пароля
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Регистрация нового пользователя
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь
	ctx := context.Background()
	_, err = rdb.Get(ctx, user.Login).Result()
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Сохраняем в Redis
	err = rdb.Set(ctx, user.Login, hashedPassword, 0).Err()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Message: "User registered successfully"})
}

// Вход в систему
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Получаем хеш пароля из Redis
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

	// Проверяем пароль
	if !checkPasswordHash(loginReq.Password, storedPassword) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": loginReq.Login,
		"exp":   time.Now().Add(time.Hour * 1).Unix(), // срок действия 1 час
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Сохраняем токен в Redis с истечением времени (кулдаун)
	err = rdb.Set(ctx, fmt.Sprintf("token:%s", loginReq.Login), tokenString, time.Hour).Err()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем токен пользователю
	json.NewEncoder(w).Encode(AuthResponse{Token: tokenString})
}

// Страница успешного входа
func successHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token not provided", http.StatusUnauthorized)
		return
	}

	// Проверяем наличие токена в Redis
	ctx := context.Background()
	storedToken, err := rdb.Get(ctx, fmt.Sprintf("token:%s", token)).Result()
	if err == redis.Nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Если токен совпадает
	if storedToken == token {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{Message: "Successfully logged in"})
	} else {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
	}
}

func main() {
	initRedis()
	defer rdb.Close()

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/success", successHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
