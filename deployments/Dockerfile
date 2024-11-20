FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY ../ ./

RUN go build -o auth-service ./cmd/auth-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=build /app/auth-service /usr/local/bin/auth-service

WORKDIR /usr/local/bin

CMD ["auth-service"]
