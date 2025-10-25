package middleware

import (
	"auth-service/internal/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		fmt.Printf("🎯 ДЕБАГ AuthMiddleware - Authorization header: '%s'\n", authHeader)

		if authHeader == "" {
			fmt.Printf("❌ ДЕБАГ AuthMiddleware - Authorization header is empty\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			c.Abort()
			return
		}

		accessToken := strings.Replace(authHeader, "Bearer ", "", 1)
		fmt.Printf("🎯 ДЕБАГ AuthMiddleware - Extracted token: '%s...'\n", accessToken[:50])

		if accessToken == "" {
			fmt.Printf("❌ ДЕБАГ AuthMiddleware - Token is empty after extraction\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный формат токена",
			})
			c.Abort()
			return
		}

		fmt.Printf("🎯 ДЕБАГ AuthMiddleware - Validating token...\n")
		claims, err := utils.ValidateToken(accessToken)
		if err != nil {
			fmt.Printf("❌ ДЕБАГ AuthMiddleware - Token validation FAILED: %v\n", err)
			fmt.Printf("❌ ДЕБАГ AuthMiddleware - Error details: %+v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Невалидный токен: " + err.Error(),
			})
			c.Abort()
			return
		}

		fmt.Printf("✅ ДЕБАГ AuthMiddleware - Token validation SUCCESS\n")
		fmt.Printf("✅ ДЕБАГ AuthMiddleware - UserID: %d, Email: %s, Role: %s\n",
			claims.UserID, claims.Email, claims.Role)

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		fmt.Printf("✅ ДЕБАГ AuthMiddleware - Context set, proceeding to handler\n")
		c.Next()
	}
}
