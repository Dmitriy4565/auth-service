# 🔐 Auth Service

Микросервис аутентификации с двухфакторной аутентификацией, JWT токенами и отправкой email через Resend.

## 🚀 Быстрый старт

### 1. Клонирование и настройка
```bash
git clone <https://github.com/Dmitriy4565/auth-service>
cd auth-service
cp .env.example .env
```
### 2. Настройка переменных окружения (.env)
```bash
# База данных
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=auth_service

# Сервер
PORT=8080
GIN_MODE=debug

# JWT (ОБЯЗАТЕЛЬНО изменить в продакшене!)
JWT_SECRET=your-super-secret-key-change-in-production

# CORS
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:5173
CORS_ALLOW_CREDENTIALS=true

# Токены
ACCESS_TOKEN_EXPIRE_MINUTES=15
REFRESH_TOKEN_EXPIRE_DAYS=7

# Resend Email Service
RESEND_API_KEY=re_your_api_key_here
RESEND_FROM_EMAIL=noreply@yourdomain.com
RESEND_FROM_NAME=Auth Service

# Клиент
CLIENT_URL=http://localhost:3000
```
### 3. Запуск в Docker
```bash
docker compose up -d --build
```
### 4. Проверка работы
```bash
curl http://localhost:8080/health
```
### 📧 Настройка Resend

### 🔴 ВАЖНО: Требования для Resend
Resend требует верифицированный домен для отправки писем в продакшене!

### Шаги настройки:
Получите API ключ на resend.com

Добавьте и верифицируйте домен в Resend Dashboard

Используйте email с верифицированного домена:

```bash
env
RESEND_FROM_EMAIL=noreply@yourcompany.com
Тестовый режим
В тестовом режиме Resend позволяет отправлять письма только на email вашего аккаунта.
```

