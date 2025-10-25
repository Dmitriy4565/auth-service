package service

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	emailService *EmailService
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: NewEmailService(),
	}
}

// TokensResponse содержит access и refresh токены
type TokensResponse struct {
	AccessToken  string
	RefreshToken string
}

// Register регистрирует нового пользователя и создает сессию верификации
func (s *AuthService) Register(registerReq *models.RegisterRequest) (*models.RegisterResponse, error) {
	// Проверяем, не существует ли уже пользователь с таким email
	existingUser, err := s.userRepo.GetUserByEmail(registerReq.Email)
	if err != nil {
		// Если это ошибка "пользователь не найден" - это нормально, продолжаем регистрацию
		if err.Error() != "пользователь не найден" {
			return nil, fmt.Errorf("ошибка проверки пользователя: %w", err)
		}
		// Пользователь не существует - продолжаем
	} else if existingUser != nil {
		// Пользователь уже существует
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

	// Генерируем activated_link и код
	activatedLink := uuid.New().String()
	code, err := utils.GenerateTwoFactorCode()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации кода: %w", err)
	}

	// Создаем сессию верификации
	session := &models.VerificationSession{
		UUID:      activatedLink,
		Email:     user.Email,
		Code:      code,
		Operation: "register",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.userRepo.CreateVerificationSession(session); err != nil {
		return nil, fmt.Errorf("ошибка создания сессии верификации: %w", err)
	}

	// Отправляем код на почту
	if err := s.emailService.Send2FACode(user.Email, code); err != nil {
		return nil, fmt.Errorf("ошибка отправки кода: %w", err)
	}

	return &models.RegisterResponse{
		Message:       "Код подтверждения отправлен на вашу почту",
		ActivatedLink: activatedLink,
	}, nil
}

// Login выполняет вход и создает сессию верификации
func (s *AuthService) Login(loginReq *models.LoginRequest) (*models.LoginResponse, error) {
	// Находим пользователя по email
	user, err := s.userRepo.GetUserByEmail(loginReq.Email)
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	// Проверяем пароль
	if !utils.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		return nil, errors.New("неверный email или пароль")
	}

	// Генерируем activated_link и код
	activatedLink := uuid.New().String()
	code, err := utils.GenerateTwoFactorCode()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации кода: %w", err)
	}

	// Создаем сессию верификации
	session := &models.VerificationSession{
		UUID:      activatedLink,
		Email:     user.Email,
		Code:      code,
		Operation: "login",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.userRepo.CreateVerificationSession(session); err != nil {
		return nil, fmt.Errorf("ошибка создания сессии верификации: %w", err)
	}

	// Отправляем код на почту
	if err := s.emailService.Send2FACode(user.Email, code); err != nil {
		return nil, fmt.Errorf("ошибка отправки кода: %w", err)
	}

	return &models.LoginResponse{
		Message:       "Код отправлен на вашу почту",
		ActivatedLink: activatedLink,
	}, nil
}

// VerifyCode проверяет код по activated_link и выдает токены
func (s *AuthService) VerifyCode(verifyReq *models.VerifyRequest) (*models.VerifyResponse, error) {
	// Находим валидную сессию верификации
	session, err := s.userRepo.GetValidVerificationSession(verifyReq.ActivatedLink, verifyReq.Code)
	if err != nil {
		return nil, errors.New("неверный или просроченный код")
	}

	// Находим пользователя по email из сессии
	user, err := s.userRepo.GetUserByEmail(session.Email)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Включаем 2FA после первой успешной проверки
	if !user.TwoFactorEnabled {
		user.TwoFactorEnabled = true
		user.TwoFactorVerified = true
		if err := s.userRepo.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("ошибка включения 2FA: %w", err)
		}
	}

	// Помечаем сессию как использованную
	if err := s.userRepo.MarkVerificationSessionAsUsed(verifyReq.ActivatedLink); err != nil {
		return nil, fmt.Errorf("ошибка при обновлении сессии: %w", err)
	}

	// Генерируем токены после успешной проверки
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &models.VerifyResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: &models.User{
			ID:                user.ID,
			Name:              user.Name,
			Lastname:          user.Lastname,
			Email:             user.Email,
			Role:              user.Role,
			TwoFactorEnabled:  user.TwoFactorEnabled,
			TwoFactorVerified: user.TwoFactorVerified,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		},
	}, nil
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

// Logout выполняет выход пользователя
func (s *AuthService) Logout(refreshToken string) error {
	return s.userRepo.DeleteSession(refreshToken)
}

// RequestResetPassword запрашивает сброс пароля
func (s *AuthService) RequestResetPassword(req *models.RequestResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		// Для безопасности не раскрываем, существует ли пользователь
		return &models.ResetPasswordResponse{
			Message: "Если пользователь с таким email существует, инструкции по сбросу пароля отправлены на почту",
		}, nil
	}

	// Генерируем UUID токен
	token := uuid.New().String()
	resetToken := &models.ResetPasswordToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Токен действителен 1 час
		Used:      false,
	}

	// Сохраняем токен в БД
	if err := s.userRepo.CreateResetPasswordToken(resetToken); err != nil {
		return nil, fmt.Errorf("ошибка создания токена сброса: %w", err)
	}

	// Формируем ссылку для сброса
	clientURL := os.Getenv("CLIENT_URL")
	if clientURL == "" {
		clientURL = "http://localhost:3000" // Значение по умолчанию
	}
	resetLink := fmt.Sprintf("%s/reset-password/%s", clientURL, token)

	// Отправляем email с ссылкой
	if err := s.emailService.SendResetPasswordEmail(user.Email, resetLink); err != nil {
		return nil, fmt.Errorf("ошибка отправки email: %w", err)
	}

	return &models.ResetPasswordResponse{
		Message: "Если пользователь с таким email существует, инструкции по сбросу пароля отправлены на почту",
	}, nil
}

// ResetPassword сбрасывает пароль используя токен
func (s *AuthService) ResetPassword(req *models.ResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	// Находим валидный токен
	resetToken, err := s.userRepo.GetValidResetToken(req.Token)
	if err != nil {
		return nil, errors.New("невалидный или просроченный токен сброса пароля")
	}

	// Находим пользователя
	user, err := s.userRepo.GetUserByID(resetToken.UserID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Хешируем новый пароль
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	// Обновляем пароль
	if err := s.userRepo.UpdateUserPassword(user.ID, hashedPassword); err != nil {
		return nil, fmt.Errorf("ошибка обновления пароля: %w", err)
	}

	// Помечаем токен как использованный
	if err := s.userRepo.MarkResetTokenAsUsed(req.Token); err != nil {
		return nil, fmt.Errorf("ошибка при обновлении токена: %w", err)
	}

	// Удаляем все сессии пользователя (выход со всех устройств)
	if err := s.userRepo.DeleteAllUserSessions(user.ID); err != nil {
		log.Printf("⚠️ Ошибка удаления сессий пользователя: %v", err)
	}

	return &models.ResetPasswordResponse{
		Message: "Пароль успешно изменен",
	}, nil
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
	s.userRepo.DeleteExpiredVerificationSessions()
	s.userRepo.DeleteExpiredResetTokens()
}
