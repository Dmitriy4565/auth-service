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

func (h *AuthHandler) setTokenCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie(
		"access_token",
		accessToken,
		15*60,
		"/",
		"",
		true,
		true,
	)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*60*60,
		"/auth/refresh",
		"",
		true,
		true,
	)
}

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

	headers := c.Writer.Header()
	fmt.Printf("📋 ДЕБАГ: Response headers before send: %v\n", headers)

	fmt.Printf("✅ ДЕБАГ: Registration successful, activated_link: %s\n", response.ActivatedLink)
	fmt.Println("🎯 ДЕБАГ: ===== REGISTER HANDLER END =====")

	c.JSON(http.StatusOK, response)
}

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

	c.JSON(http.StatusOK, gin.H{
		"access_token":  response.AccessToken,
		"refresh_token": response.RefreshToken})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	fmt.Printf("🎯 ДЕБАГ Profile - user_id from context: %v, exists: %v\n", userID, exists)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Требуется авторизация - user_id not found in context",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Пользователь не найден",
		})
		return
	}

	profile := models.ProfileResponse{
		ID:       user.ID,
		Name:     user.Name,
		Lastname: user.Lastname,
		Email:    user.Email,
		Role:     user.Role, // 👈 ДОБАВЬ ЭТУ СТРОКУ
	}

	c.JSON(http.StatusOK, profile)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует refresh token в заголовке Authorization"})
		return
	}

	refreshToken := strings.Replace(authHeader, "Bearer ", "", 1)
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат refresh token"})
		return
	}

	tokens, err := h.authService.RefreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"message":       "Токены успешно обновлены",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteNoneMode)

	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Успешный выход",
	})
}

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

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Невалидный токен",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
