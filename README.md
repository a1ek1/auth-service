# Auth Service

Это сервис для регистрации, логина и проверки токенов пользователей с использованием JWT и Redis.

## Запуск проекта

### Требования

Для запуска сервиса вам понадобятся:

- Docker
- Docker Compose

### Запуск с помощью Docker

1. Клонируйте репозиторий и перейдите в каталог проекта:

   ```bash
   git clone https://github.com/a1ek1/auth-service.git
   cd auth-service
   ```

2. Соберите и запустите проект с помощью Docker Compose

   ```bash
   cd deployments
   docker-compose up --build
   ```

3. После успешного запуска, сервис будет доступен на порту 8080 на вашем локальном компьютере.

4. Для остановки вам нужно ввести 

   ```bash
   docker-compose down
   ```
   или 

   ```bash
   docker stop $(docker ps -q)
   docker rm $(docker ps -aq)
   ```

### Регистрация пользователя

Для регистрации нового пользователя выполните следующий запрос:

```bash
curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d "{\"login\": \"oleg\", \"password\": \"password123\"}"
```

### Авторизация пользователя
Для авторизации пользователя выполните следующий запрос:
```bash
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d "{\"login\": \"oleg\", \"password\": \"password123\"}" -i
```

### Проверка доступа к данным
Для доступа к защищённым данным, используйте следующий запрос, заменив {token} на ваш токен, полученный при авторизации:
```bash
curl -X GET http://localhost:8080/success -H "Authorization: Bearer {token}
```
