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

type TokensResponse struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(registerReq *models.RegisterRequest) (*models.RegisterResponse, error) {
	existingUser, err := s.userRepo.GetUserByEmail(registerReq.Email)
	if err != nil {
		if err.Error() != "пользователь не найден" {
			return nil, fmt.Errorf("ошибка проверки пользователя: %w", err)
		}
	} else if existingUser != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	hashedPassword, err := utils.HashPassword(registerReq.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	user := &models.User{
		Name:         registerReq.Name,
		Lastname:     registerReq.Lastname,
		Email:        registerReq.Email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	activatedLink := uuid.New().String()
	code, err := utils.GenerateTwoFactorCode()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации кода: %w", err)
	}

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

	if err := s.emailService.Send2FACode(user.Email, code); err != nil {
		return nil, fmt.Errorf("ошибка отправки кода: %w", err)
	}

	return &models.RegisterResponse{
		Message:       "Код подтверждения отправлен на вашу почту",
		ActivatedLink: activatedLink,
	}, nil
}

func (s *AuthService) Login(loginReq *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(loginReq.Email)
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	if !utils.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		return nil, errors.New("неверный email или пароль")
	}

	activatedLink := uuid.New().String()
	code, err := utils.GenerateTwoFactorCode()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации кода: %w", err)
	}

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

	if err := s.emailService.Send2FACode(user.Email, code); err != nil {
		return nil, fmt.Errorf("ошибка отправки кода: %w", err)
	}

	return &models.LoginResponse{
		Message:       "Код отправлен на вашу почту",
		ActivatedLink: activatedLink,
	}, nil
}

func (s *AuthService) VerifyCode(verifyReq *models.VerifyRequest) (*models.VerifyResponse, error) {
	session, err := s.userRepo.GetValidVerificationSession(verifyReq.ActivatedLink, verifyReq.Code)
	if err != nil {
		return nil, errors.New("неверный или просроченный код")
	}

	user, err := s.userRepo.GetUserByEmail(session.Email)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if !user.TwoFactorEnabled {
		user.TwoFactorEnabled = true
		user.TwoFactorVerified = true
		if err := s.userRepo.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("ошибка включения 2FA: %w", err)
		}
	}

	if err := s.userRepo.MarkVerificationSessionAsUsed(verifyReq.ActivatedLink); err != nil {
		return nil, fmt.Errorf("ошибка при обновлении сессии: %w", err)
	}

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

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}

func (s *AuthService) RefreshTokens(refreshToken string) (*TokensResponse, error) {
	session, err := s.userRepo.GetSessionByToken(refreshToken)
	if err != nil {
		return nil, errors.New("невалидный refresh token")
	}

	user, err := s.userRepo.GetUserByID(session.UserID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if err := s.userRepo.DeleteSession(refreshToken); err != nil {
		return nil, fmt.Errorf("ошибка удаления сессии: %w", err)
	}

	return s.generateTokens(user)
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.userRepo.DeleteSession(refreshToken)
}

func (s *AuthService) RequestResetPassword(req *models.RequestResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return &models.ResetPasswordResponse{
			Message: "Если пользователь с таким email существует, инструкции по сбросу пароля отправлены на почту",
		}, nil
	}

	token := uuid.New().String()
	resetToken := &models.ResetPasswordToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	if err := s.userRepo.CreateResetPasswordToken(resetToken); err != nil {
		return nil, fmt.Errorf("ошибка создания токена сброса: %w", err)
	}

	clientURL := os.Getenv("CLIENT_URL")
	if clientURL == "" {
		clientURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/auth/reset-password/%s", clientURL, token)

	if err := s.emailService.SendResetPasswordEmail(user.Email, resetLink); err != nil {
		return nil, fmt.Errorf("ошибка отправки email: %w", err)
	}

	return &models.ResetPasswordResponse{
		Message: "Если пользователь с таким email существует, инструкции по сбросу пароля отправлены на почту",
	}, nil
}

func (s *AuthService) ResetPassword(req *models.ResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	resetToken, err := s.userRepo.GetValidResetToken(req.Token)
	if err != nil {
		return nil, errors.New("невалидный или просроченный токен сброса пароля")
	}

	user, err := s.userRepo.GetUserByID(resetToken.UserID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	if err := s.userRepo.UpdateUserPassword(user.ID, hashedPassword); err != nil {
		return nil, fmt.Errorf("ошибка обновления пароля: %w", err)
	}

	if err := s.userRepo.MarkResetTokenAsUsed(req.Token); err != nil {
		return nil, fmt.Errorf("ошибка при обновлении токена: %w", err)
	}

	if err := s.userRepo.DeleteAllUserSessions(user.ID); err != nil {
		log.Printf("⚠️ Ошибка удаления сессий пользователя: %v", err)
	}

	return &models.ResetPasswordResponse{
		Message: "Пароль успешно изменен",
	}, nil
}

func (s *AuthService) generateTokens(user *models.User) (*TokensResponse, error) {
	accessToken, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации refresh token: %w", err)
	}

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.userRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("ошибка создания сессии: %w", err)
	}

	go s.cleanupExpiredData()

	return &TokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) cleanupExpiredData() {
	s.userRepo.DeleteExpiredSessions()
	s.userRepo.DeleteExpiredTwoFactorCodes()
	s.userRepo.DeleteExpiredVerificationSessions()
	s.userRepo.DeleteExpiredResetTokens()
}
