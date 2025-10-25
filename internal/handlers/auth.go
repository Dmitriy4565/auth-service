package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/service"
	"auth-service/internal/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
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

// VerifyEmail обрабатывает проверку кода верификации
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

	// Устанавливаем access token в httpOnly cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		response.AccessToken,
		3600, // 1 час
		"/",
		"",
		false,
		true,
	)

	// Возвращаем пользователя и refresh token в теле ответа
	c.JSON(http.StatusOK, gin.H{
		"refresh_token": response.RefreshToken,
		"user":          response.User,
		"message":       "Верификация успешно завершена",
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
	// Получаем refresh token из заголовка Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует refresh token"})
		return
	}

	// Извлекаем токен из заголовка (формат: "Bearer {token}")
	refreshToken := strings.Replace(authHeader, "Bearer ", "", 1)
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
		return
	}

	tokens, err := h.authService.RefreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем новый access token в cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		tokens.AccessToken,
		3600, // 1 час
		"/",
		"",
		false,
		true,
	)

	// Возвращаем новый refresh token в теле ответа
	c.JSON(http.StatusOK, gin.H{
		"refresh_token": tokens.RefreshToken,
		"message":       "Токены успешно обновлены",
	})
}

// Logout обрабатывает выход пользователя
func (h *AuthHandler) Logout(c *gin.Context) {
	// Получаем refresh token из заголовка Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Отсутствует токен"})
		return
	}

	refreshToken := strings.Replace(authHeader, "Bearer ", "", 1)
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат токена"})
		return
	}

	// Удаляем refresh token из базы
	if err := h.authService.Logout(refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выходе из системы"})
		return
	}

	// Очищаем access token cookie
	c.SetCookie(
		"access_token",
		"",
		-1, // удаляем cookie
		"/",
		"",
		false,
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
