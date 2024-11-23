### Регистрация пользователя

Для регистрации нового пользователя выполните следующий запрос:

```bash
curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d "{\"username\": \"user1\", \"password\": \"password123\"}"
```

### Авторизация пользователя
Для авторизации пользователя выполните следующий запрос:
```bash
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d "{\"username\": \"user1\", \"password\": \"password123\"}"
```

### Проверка доступа к данным
Выполните следующий запрос:
```bash
curl -X POST http://localhost:8080/success?token="{{ваш_токен}}"
```
