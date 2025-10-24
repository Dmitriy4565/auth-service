package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/service"
	"auth-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register обрабатывает регистрацию пользователя
// @Summary Регистрация пользователя
// @Description Создает нового пользователя и возвращает токены в cookies
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} models.TokenResponse "Токены установлены в cookies"
// @Failure 400 {object} gin.H "Ошибка валидации данных"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var registerReq models.RegisterRequest

	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные: " + err.Error(),
		})
		return
	}

	// Регистрируем пользователя
	user, err := h.authService.Register(&registerReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Генерируем токены для нового пользователя
	tokens, err := h.authService.GenerateTokensForUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка генерации токенов",
		})
		return
	}

	// Устанавливаем токены в httpOnly cookies
	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)

	c.JSON(http.StatusOK, models.TokenResponse{
		Message: "Пользователь успешно зарегистрирован",
	})
}

// Login обрабатывает вход пользователя
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя и возвращает токены в cookies
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.TokenResponse "Токены установлены в cookies"
// @Failure 400 {object} gin.H "Ошибка валидации данных"
// @Failure 401 {object} gin.H "Неверные учетные данные"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные: " + err.Error(),
		})
		return
	}

	// Аутентифицируем пользователя
	tokens, err := h.authService.Login(&loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Устанавливаем токены в httpOnly cookies
	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)

	c.JSON(http.StatusOK, models.TokenResponse{
		Message: "Успешный вход в систему",
	})
}

// Profile возвращает профиль пользователя
// @Summary Профиль пользователя
// @Description Возвращает данные текущего пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.ProfileResponse "Данные пользователя"
// @Failure 401 {object} gin.H "Требуется авторизация"
// @Router /auth/profile [get]
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
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.TokenResponse "Токены обновлены"
// @Failure 401 {object} gin.H "Невалидный refresh token"
// @Router /auth/refresh [get]
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Получаем refresh token из cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Refresh token не найден",
		})
		return
	}

	// Обновляем токены
	tokens, err := h.authService.RefreshTokens(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Устанавливаем новые токены в cookies
	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)

	c.JSON(http.StatusOK, models.TokenResponse{
		Message: "Токены успешно обновлены",
	})
}

// setAuthCookies устанавливает access и refresh токены в httpOnly cookies
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	// Access token на 15 минут
	c.SetCookie("access_token", accessToken, 15*60, "/", "", false, true)
	// Refresh token на 7 дней
	c.SetCookie("refresh_token", refreshToken, 7*24*60*60, "/", "", false, true)

	// Также возвращаем access token в теле ответа для удобства фронта
	c.Header("X-Access-Token", accessToken)
}

// Logout выполняет выход пользователя
// @Summary Выход из системы
// @Description Выполняет logout и удаляет токены
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.TokenResponse "Успешный выход"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Удаляем токены из cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, models.TokenResponse{
		Message: "Успешный выход из системы",
	})
}
