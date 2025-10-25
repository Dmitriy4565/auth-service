package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/pkg/database"
	"fmt"
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
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.GET("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/requestToResetPassword", authHandler.RequestResetPassword)
		auth.POST("/resetPassword", authHandler.ResetPassword)
	}

	// Защищенные маршруты (требуют авторизации)
	protected := router.Group("/auth")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", authHandler.Profile)
	}

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

	// 🔥 ТЕСТОВЫЙ ЭНДПОИНТ БЕЗ ВСЯКОЙ ЛОГИКИ
	router.GET("/test-cookies", func(c *gin.Context) {
		fmt.Println("🔍 ТЕСТ: /test-cookies вызван")

		// Проверяем входящие куки
		cookies := c.Request.Cookies()
		fmt.Printf("🔍 ТЕСТ: Входящие куки: %v\n", cookies)

		// Просто возвращаем текст
		c.JSON(200, gin.H{
			"message":     "Это тестовый endpoint",
			"has_cookies": len(cookies) > 0,
		})
	})

	router.POST("/test-register", func(c *gin.Context) {
		fmt.Println("🔍 ТЕСТ: /test-register вызван")

		// Принудительно очищаем ВСЕ возможные куки
		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.SetCookie("session", "", -1, "/", "", false, true)

		c.JSON(200, gin.H{
			"message": "Тестовая регистрация - куки очищены",
		})
	})
}
