package repository

import (
	"auth-service/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) CreateSession(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *UserRepository) GetSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("refresh_token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	return &session, err
}

func (r *UserRepository) DeleteSession(token string) error {
	return r.db.Where("refresh_token = ?", token).Delete(&models.Session{}).Error
}

func (r *UserRepository) DeleteExpiredSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

func (r *UserRepository) DeleteAllUserSessions(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func (r *UserRepository) CreateTwoFactorCode(code *models.TwoFactorCode) error {
	return r.db.Create(code).Error
}

func (r *UserRepository) GetValidTwoFactorCode(userID uint, code string) (*models.TwoFactorCode, error) {
	var twoFactorCode models.TwoFactorCode
	err := r.db.Where("user_id = ? AND code = ? AND used = ? AND expires_at > ?", userID, code, false, time.Now()).First(&twoFactorCode).Error
	return &twoFactorCode, err
}

func (r *UserRepository) MarkTwoFactorCodeAsUsed(id uint) error {
	return r.db.Model(&models.TwoFactorCode{}).Where("id = ?", id).Update("used", true).Error
}

func (r *UserRepository) DeleteExpiredTwoFactorCodes() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.TwoFactorCode{}).Error
}

func (r *UserRepository) CreateVerificationSession(session *models.VerificationSession) error {
	return r.db.Create(session).Error
}

func (r *UserRepository) GetValidVerificationSession(uuid, code string) (*models.VerificationSession, error) {
	var session models.VerificationSession
	err := r.db.Where("uuid = ? AND code = ? AND used = ? AND expires_at > ?", uuid, code, false, time.Now()).First(&session).Error
	return &session, err
}

func (r *UserRepository) MarkVerificationSessionAsUsed(uuid string) error {
	return r.db.Model(&models.VerificationSession{}).Where("uuid = ?", uuid).Update("used", true).Error
}

func (r *UserRepository) DeleteExpiredVerificationSessions() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.VerificationSession{}).Error
}

func (r *UserRepository) CreateResetPasswordToken(token *models.ResetPasswordToken) error {
	return r.db.Create(token).Error
}

func (r *UserRepository) GetValidResetToken(token string) (*models.ResetPasswordToken, error) {
	var resetToken models.ResetPasswordToken
	err := r.db.Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).First(&resetToken).Error
	return &resetToken, err
}

func (r *UserRepository) MarkResetTokenAsUsed(token string) error {
	return r.db.Model(&models.ResetPasswordToken{}).Where("token = ?", token).Update("used", true).Error
}

func (r *UserRepository) UpdateUserPassword(userID uint, newPasswordHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", newPasswordHash).Error
}

func (r *UserRepository) DeleteExpiredResetTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.ResetPasswordToken{}).Error
}
