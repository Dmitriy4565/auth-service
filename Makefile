# Makefile для управления микросервисом аутентификации

.PHONY: build run test clean migrate docker-up docker-down

# Сборка приложения
build:
	@echo "🔨 Сборка приложения..."
	go build -o bin/main ./cmd/server

# Запуск приложения
run: build
	@echo "🚀 Запуск приложения..."
	./bin/main

# Запуск тестов
test:
	@echo "🧪 Запуск тестов..."
	go test ./... -v

# Очистка билдов
clean:
	@echo "🧹 Очистка..."
	rm -rf bin/

# Запуск в Docker
docker-up:
	@echo "🐳 Запуск Docker контейнеров..."
	docker-compose up -d

# Остановка Docker
docker-down:
	@echo "🛑 Остановка Docker контейнеров..."
	docker-compose down

# Просмотр логов
logs:
	docker-compose logs -f auth-service

# Миграции базы данных
migrate:
	@echo "🗃️  Выполнение миграций..."
	# Здесь можно добавить команды для миграций

# Запуск в режиме разработки
dev:
	@echo "👨‍💻 Запуск в режиме разработки..."
	air

# Генерация документации
docs:
	@echo "📚 Генерация документации..."
	# Здесь можно добавить swagger или другую документацию