package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/service"
	"auth-service/internal/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// setTokenCookies устанавливает access и refresh токены в httpOnly куки
func (h *AuthHandler) setTokenCookies(c *gin.Context, accessToken, refreshToken string) {
	// Access Token кука (15 минут, доступен для всех API endpoints)
	c.SetCookie(
		"access_token",
		accessToken,
		15*60, // 15 минут
		"/",
		"",
		true, // Secure (true для продакшена)
		true, // HttpOnly
	)

	// Refresh Token кука (7 дней, доступен только для refresh endpoint)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*60*60, // 7 дней
		"/auth/refresh",
		"",
		true, // Secure
		true, // HttpOnly
	)
}

// clearTokenCookies очищает токены из кук
func (h *AuthHandler) clearTokenCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/auth/refresh", "", true, true)
}

func (h *AuthHandler) Register(c *gin.Context) {
	fmt.Println("🎯 ДЕБАГ: ===== REGISTER HANDLER START =====")

	var registerReq models.RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		fmt.Printf("❌ ДЕБАГ: Validation error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные: " + err.Error(),
		})
		return
	}

	fmt.Printf("📧 ДЕБАГ: Registering user: %s\n", registerReq.Email)

	response, err := h.authService.Register(&registerReq)
	if err != nil {
		fmt.Printf("❌ ДЕБАГ: Service error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Проверяем заголовки ПЕРЕД отправкой ответа
	headers := c.Writer.Header()
	fmt.Printf("📋 ДЕБАГ: Response headers before send: %v\n", headers)

	fmt.Printf("✅ ДЕБАГ: Registration successful, activated_link: %s\n", response.ActivatedLink)
	fmt.Println("🎯 ДЕБАГ: ===== REGISTER HANDLER END =====")

	c.JSON(http.StatusOK, response)
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные: " + err.Error(),
		})
		return
	}

	response, err := h.authService.Login(&loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// // VerifyEmail обрабатывает проверку кода верификации
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req models.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	response, err := h.authService.VerifyCode(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// SameSite=None но БЕЗ Secure
	c.SetSameSite(http.SameSiteNoneMode)

	// Устанавливаем access token в httpOnly cookie
	c.SetCookie(
		"access_token",
		response.AccessToken,
		3600, // 1 час
		"/",
		"",    // домен
		false, // 🔥 Secure = false для HTTP разработки
		true,  // httpOnly
	)

	c.SetCookie(
		"refresh_token",
		response.RefreshToken,
		7*24*3600, // 7 дней
		"/",
		"",    // домен
		false, // 🔥 Secure = false для HTTP разработки
		true,  // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Верификация успешно завершена",
	})
}

// Profile возвращает профиль пользователя
func (h *AuthHandler) Profile(c *gin.Context) {
	// Получаем access token из cookie
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Требуется авторизация",
		})
		return
	}

	// Валидируем токен и получаем данные пользователя
	claims, err := utils.ValidateToken(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Невалидный токен",
		})
		return
	}

	// Находим пользователя в БД
	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Пользователь не найден",
		})
		return
	}

	// Возвращаем профиль пользователя
	profile := models.ProfileResponse{
		Name:     user.Name,
		Lastname: user.Lastname,
		Email:    user.Email,
	}

	c.JSON(http.StatusOK, profile)
}

// Refresh обновляет токены
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует refresh token"})
		return
	}

	tokens, err := h.authService.RefreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 🔥 СТАВИМ SameSite=None
	c.SetSameSite(http.SameSiteNoneMode)

	// Устанавливаем новый access token в cookie
	c.SetCookie(
		"access_token",
		tokens.AccessToken,
		3600,
		"/",
		"",
		true, // 🔥 secure = true
		true,
	)

	c.SetCookie(
		"refresh_token",
		tokens.RefreshToken,
		7*24*3600,
		"/",
		"",
		true, // 🔥 secure = true
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Токены успешно обновлены",
	})
}

// Logout обрабатывает выход пользователя
func (h *AuthHandler) Logout(c *gin.Context) {
	// 🔥 СТАВИМ SameSite=None
	c.SetSameSite(http.SameSiteNoneMode)

	// Очищаем access token cookie
	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		true, // 🔥 secure = true
		true,
	)

	// Очищаем refresh token cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		true, // 🔥 secure = true
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Успешный выход",
	})
}

// RequestResetPassword обрабатывает запрос на сброс пароля
func (h *AuthHandler) RequestResetPassword(c *gin.Context) {
	var req models.RequestResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	response, err := h.authService.RequestResetPassword(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResetPassword обрабатывает сброс пароля
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса"})
		return
	}

	response, err := h.authService.ResetPassword(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AuthMiddleware middleware для проверки access token
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем access token из куки
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

		// Сохраняем userID в контекст для использования в обработчиках
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
