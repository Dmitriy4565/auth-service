# Используем официальный образ Go для сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы модулей для кэширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код в контейнер
COPY . .

# Собираем приложение в статический бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Второй этап: создаем легковесный образ для запуска
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя app для безопасности (не запускаем от root)
RUN addgroup -S app && adduser -S app -G app

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Копируем .env.example (в продакшне использовать реальные env переменные)
COPY --from=builder /app/.env.example .env.example

# Меняем владельца файлов на пользователя app
RUN chown -R app:app ./

# Переключаемся на пользователя app
USER app

# Открываем порт, который будет использовать приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]