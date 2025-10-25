package models

import (
	"time"
)

// User - основная модель пользователя
type User struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"size:100;not null" json:"name"`
	Lastname          string    `gorm:"size:100;not null" json:"lastname"`
	Email             string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash      string    `gorm:"size:255;not null" json:"-"`
	Role              string    `gorm:"size:50;not null;default:user" json:"role"`
	TwoFactorEnabled  bool      `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret   string    `gorm:"size:255" json:"-"`
	TwoFactorVerified bool      `gorm:"default:false" json:"two_factor_verified"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Session - модель сессии для refresh токенов
type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	RefreshToken string    `gorm:"size:255;uniqueIndex;not null" json:"-"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// TwoFactorCode - модель для кодов двухфакторной аутентификации
type TwoFactorCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// VerificationSession - модель для сессий верификации
type VerificationSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UUID      string    `gorm:"size:36;uniqueIndex;not null" json:"activated_link"`
	Email     string    `gorm:"size:255;not null" json:"email"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	Operation string    `gorm:"size:20;not null" json:"operation"` // "register" или "login"
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// ResetPasswordToken - модель для токена сброса пароля
type ResetPasswordToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Token     string    `gorm:"size:255;uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// DTO для запросов и ответов API
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Lastname string `json:"lastname" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=5"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=5"`
}

type VerifyRequest struct {
	ActivatedLink string `json:"activated_link" binding:"required"`
	Code          string `json:"code" binding:"required,len=6"`
}

type RegisterResponse struct {
	Message       string `json:"message"`
	ActivatedLink string `json:"activated_link"`
}

type LoginResponse struct {
	Message       string `json:"message"`
	ActivatedLink string `json:"activated_link"`
}

type VerifyResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

type ProfileResponse struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
}

type TokenResponse struct {
	Message string `json:"message"`
}

// DTO для запросов сброса пароля
type RequestResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=5"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// Добавьте в конец файла с моделями:

// Verify2FARequest запрос для верификации 2FA кода
type Verify2FARequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// Verify2FAResponse ответ после успешной верификации 2FA
type Verify2FAResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}
