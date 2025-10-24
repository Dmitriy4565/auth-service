package repository

import (
	"auth-service/internal/models"
	"time"

	"gorm.io/gorm"
)

// UserRepository предоставляет методы для работы с пользователями в БД
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser создает нового пользователя в базе данных
func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail находит пользователя по email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID находит пользователя по ID
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser обновляет данные пользователя
func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// CreateSession создает новую сессию для refresh token
func (r *UserRepository) CreateSession(session *models.Session) error {
	return r.db.Create(session).Error
}

// GetSessionByToken находит активную сессию по refresh token
func (r *UserRepository) GetSessionByToken(refreshToken string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("refresh_token = ? AND expires_at > ?", refreshToken, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteSession удаляет сессию по refresh token
func (r *UserRepository) DeleteSession(refreshToken string) error {
	return r.db.Where("refresh_token = ?", refreshToken).Delete(&models.Session{}).Error
}

// DeleteExpiredSessions удаляет все просроченные сессии
func (r *UserRepository) DeleteExpiredSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

// CreateTwoFactorCode создает временный код для двухфакторной аутентификации
func (r *UserRepository) CreateTwoFactorCode(tfCode *models.TwoFactorCode) error {
	return r.db.Create(tfCode).Error
}

// Добавляем методы для работы с VerificationSession
func (r *UserRepository) CreateVerificationSession(session *models.VerificationSession) error {
	return r.db.Create(session).Error
}

func (r *UserRepository) GetValidVerificationSession(uuid, code string) (*models.VerificationSession, error) {
	var session models.VerificationSession
	err := r.db.Where("uuid = ? AND code = ? AND used = false AND expires_at > ?",
		uuid, code, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UserRepository) MarkVerificationSessionAsUsed(uuid string) error {
	return r.db.Model(&models.VerificationSession{}).Where("uuid = ?", uuid).Update("used", true).Error
}

func (r *UserRepository) DeleteExpiredVerificationSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.VerificationSession{}).Error
}

// MarkTwoFactorCodeAsUsed помечает код как использованный
func (r *UserRepository) MarkTwoFactorCodeAsUsed(tfCodeID uint) error {
	return r.db.Model(&models.TwoFactorCode{}).Where("id = ?", tfCodeID).Update("used", true).Error
}

// DeleteExpiredTwoFactorCodes удаляет все просроченные коды 2FA
func (r *UserRepository) DeleteExpiredTwoFactorCodes() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.TwoFactorCode{}).Error
}
