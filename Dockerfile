FROM golang:1.25-alpine

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache gcc musl-dev

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -ldflags="-s -w" -o auth-service ./cmd/server

# Экспортируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./auth-service"]