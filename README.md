# Auth Service

Это сервис для регистрации, логина и проверки токенов пользователей с использованием JWT и Redis.

## Запуск проекта c помощью Docker Compose

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


## Запуск с помощью Minikube и Docker

### Требования

1. Убедитесь, что Docker установлен и запущен на вашем компьютере
2. Установите Minicube:
```bash
   curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
   sudo install minikube-linux-amd64 /usr/local/bin/minikube
```
3. Установите kubectl
```bash
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

## Запуск с помощью Minikube

1. Клонируйте репозиторий и перейдите в каталог проекта:

```bash
   git clone https://github.com/a1ek1/auth-service.git
   cd auth-service
```

2. Запустите Minikube:
```bash
    minikube start --driver=docker 
```
3. Настройте использование Minikube Docker-демона:
```bash
    eval $(minikube docker-env) 
```
4. Примените все манифесты Kubernetes:
```bash
    kubectl apply -f deployments/k8s/
```
5. Убедитесь, что поды запущены:
```bash
    kubectl get pods
```
6. Получите URL для доступа к сервису:
```bash
    minikube service auth-service --url
```
### Регистрация пользователя

Для регистрации нового пользователя выполните следующий запрос:

```bash
   curl -X POST <URL>/register -H "Content-Type: application/json" -d "{\"login\": \"oleg\", \"password\": \"password123\"}"
```

### Авторизация пользователя
Для авторизации пользователя выполните следующий запрос:

```bash
   curl -X POST <URL>/login -H "Content-Type: application/json" -d "{\"login\": \"oleg\", \"password\": \"password123\"}" -i
```

### Проверка доступа к данным
Для доступа к защищённым данным, используйте следующий запрос, заменив {token} на ваш токен, полученный при авторизации:

```bash
   curl -X GET <URL>/success -H "Authorization: Bearer {token}
```