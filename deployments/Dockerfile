FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

COPY . ./
RUN go build -o auth-service ./cmd/auth-service

FROM alpine:latest

# Установка зависимостей
RUN apk --no-cache add ca-certificates bash curl

# Скачиваем утилиту golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xz -C /usr/local/bin

# Копируем необходимые файлы
COPY --from=build /app/auth-service /usr/local/bin/auth-service
COPY .env /usr/local/bin/.env
COPY ./migrations /usr/local/bin/migrations

WORKDIR /usr/local/bin

# Выполнение миграций и запуск основного приложения
CMD ["sh", "-c", "migrate -path=/usr/local/bin/migrations -database 'postgres://auth_user:auth_password@postgres:5432/auth_db?sslmode=disable' up && auth-service"]
