# 🔐 Auth Service

Микросервис аутентификации с двухфакторной аутентификацией, JWT токенами и отправкой email через Resend.

## 🏗️ Структура проекта
auth-service/
├── cmd/
│ └── server/
│ └── main.go # Точка входа приложения
├── internal/
│ ├── config/
│ │ └── config.go # Конфигурация приложения
│ ├── handlers/
│ │ └── auth.go # HTTP обработчики
│ ├── middleware/
│ │ └── auth.go # JWT middleware
│ ├── models/
│ │ └── user.go # Модели данных
│ ├── repository/
│ │ └── postgres.go # Репозиторий БД
│ ├── service/
│ │ ├── auth.go # Бизнес-логика аутентификации
│ │ └── email.go # Сервис отправки email (Resend)
│ └── utils/
│ ├── jwt.go # JWT утилиты
│ └── two_factor.go # 2FA утилиты
├── pkg/
│ └── database/
│ └── postgres.go # Подключение к PostgreSQL
├── migrations/
│ ├── 001_init.sql # Инициализация БД
│ └── 002_add_reset_password_tokens.sql # Токены сброса пароля
├── docker-compose.yml # Docker окружение
├── Dockerfile # Сборка Go приложения
├── deploy.py # Скрипт деплоя
└── connect.py # Подключение к серверу

## 🚀 Быстрый старт

### 1. Клонирование и настройка
```bash
git clone <https://github.com/Dmitriy4565/auth-service>
cd auth-service
cp .env.example .env
