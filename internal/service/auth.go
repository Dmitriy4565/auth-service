package service

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"errors"
	"fmt"
	"time"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// TokensResponse содержит access и refresh токены
type TokensResponse struct {
	AccessToken  string
	RefreshToken string
}

// GenerateTokensForUser генерирует токены для существующего пользователя
func (s *AuthService) GenerateTokensForUser(user *models.User) (*TokensResponse, error) {
	return s.generateTokens(user)
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(registerReq *models.RegisterRequest) (*models.User, error) {
	// Проверяем, не существует ли уже пользователь с таким email
	existingUser, _ := s.userRepo.GetUserByEmail(registerReq.Email)
	if existingUser != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := utils.HashPassword(registerReq.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	// Создаем пользователя
	user := &models.User{
		Name:         registerReq.Name,
		Lastname:     registerReq.Lastname,
		Email:        registerReq.Email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	// Сохраняем пользователя
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return user, nil
}

// Login выполняет аутентификацию пользователя
func (s *AuthService) Login(loginReq *models.LoginRequest) (*TokensResponse, error) {
	// Находим пользователя по email
	user, err := s.userRepo.GetUserByEmail(loginReq.Email)
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	// Проверяем пароль
	if !utils.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		return nil, errors.New("неверный email или пароль")
	}

	// Генерируем токены
	return s.generateTokens(user)
}

// GetUserByID возвращает пользователя по ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}

// RefreshTokens обновляет access и refresh токены
func (s *AuthService) RefreshTokens(refreshToken string) (*TokensResponse, error) {
	// Находим активную сессию
	session, err := s.userRepo.GetSessionByToken(refreshToken)
	if err != nil {
		return nil, errors.New("невалидный refresh token")
	}

	// Находим пользователя
	user, err := s.userRepo.GetUserByID(session.UserID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Удаляем старую сессию
	if err := s.userRepo.DeleteSession(refreshToken); err != nil {
		return nil, fmt.Errorf("ошибка удаления сессии: %w", err)
	}

	// Генерируем новые токены
	return s.generateTokens(user)
}

// generateTokens создает access и refresh tokens для пользователя
func (s *AuthService) generateTokens(user *models.User) (*TokensResponse, error) {
	// Генерируем access token
	accessToken, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации access token: %w", err)
	}

	// Генерируем refresh token
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации refresh token: %w", err)
	}

	// Создаем сессию для refresh token (7 дней)
	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.userRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("ошибка создания сессии: %w", err)
	}

	// В фоне очищаем просроченные данные
	go s.cleanupExpiredData()

	return &TokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// cleanupExpiredData очищает просроченные сессии
func (s *AuthService) cleanupExpiredData() {
	s.userRepo.DeleteExpiredSessions()
	s.userRepo.DeleteExpiredTwoFactorCodes()
}
