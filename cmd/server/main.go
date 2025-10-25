package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/database"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден, используются переменные окружения по умолчанию")
	}

	// Загружаем конфигурацию
	cfg := config.Load()
	log.Printf("📁 Конфигурация загружена: БД=%s, Порт=%s", cfg.DBName, cfg.Port)

	// Подключаемся к базе данных
	db, err := database.NewPostgresDB(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к базе данных:", err)
	}

	// Инициализируем слои приложения
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Создаем Gin роутер
	router := gin.Default()

	// CORS middleware для фронта
	router.Use(func(c *gin.Context) {
		allowedOrigins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
		origin := c.Request.Header.Get("Origin")

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
				break
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Глобальные мидлвари
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Публичные маршруты аутентификации
	// В файле с роутами добавьте:

	// Защищенные роуты с middleware
	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware()) // ← ДОБАВИТЬ ЭТУ СТРОЧКУ
	{
		protected.GET("/profile", authHandler.Profile)
		// другие защищенные эндпоинты...
	}

	// Auth роуты
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/refresh", authHandler.Refresh) // ← Refresh теперь через куки
		auth.POST("/logout", authHandler.Logout)   // ← Logout теперь через куки
		auth.POST("/request-reset-password", authHandler.RequestResetPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}
	// Тестовый эндпоинт для проверки кук
	router.GET("/auth/test-cookies", func(c *gin.Context) {
		// Получаем куки
		accessToken, _ := c.Cookie("access_token")
		refreshToken, _ := c.Cookie("refresh_token")

		hasAccess := accessToken != ""
		hasRefresh := refreshToken != ""

		c.JSON(200, gin.H{
			"has_access_token":     hasAccess,
			"has_refresh_token":    hasRefresh,
			"access_token_length":  len(accessToken),
			"refresh_token_length": len(refreshToken),
			"message":              "Этот эндпоинт проверяет куки",
		})
	})

	// Маршрут для проверки здоровья
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "auth-service",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// Запускаем сервер
	log.Printf("✅ Сервер запущен на порту %s", cfg.Port)
	log.Printf("📚 API доступно по http://localhost:%s", cfg.Port)
	log.Printf("🌐 CORS разрешены для: %s", os.Getenv("CORS_ALLOW_ORIGINS"))

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Ошибка запуска сервера:", err)
	}
}
