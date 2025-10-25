package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
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
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден, используются переменные окружения по умолчанию")
	}

	cfg := config.Load()
	log.Printf("📁 Конфигурация загружена: БД=%s, Порт=%s", cfg.DBName, cfg.Port)

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

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	router := gin.Default()

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

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/request-reset-password", authHandler.RequestResetPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	protected := router.Group("/auth")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", authHandler.Profile)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "auth-service",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	log.Printf("Сервер запущен на порту %s", cfg.Port)
	log.Printf("API доступно по http://localhost:%s", cfg.Port)
	log.Printf("CORS разрешены для: %s", os.Getenv("CORS_ALLOW_ORIGINS"))

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Ошибка запуска сервера:", err)
	}
}
