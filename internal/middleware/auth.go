package middleware

import (
	"auth-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет access token из cookie
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем access token из cookie
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			c.Abort()
			return
		}

		// Валидируем токен
		claims, err := utils.ValidateToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Невалидный токен",
			})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контекст
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}
