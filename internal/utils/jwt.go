package utils

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// GetTokenExpiration возвращает время жизни токенов
func GetTokenExpiration() (accessExp time.Duration, refreshExp time.Duration) {
	accessMinutes, _ := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES"))
	refreshDays, _ := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRE_DAYS"))

	if accessMinutes == 0 {
		accessMinutes = 15 // 15 минут по умолчанию
	}
	if refreshDays == 0 {
		refreshDays = 7 // 7 дней по умолчанию
	}

	return time.Duration(accessMinutes) * time.Minute,
		time.Duration(refreshDays) * 24 * time.Hour
}

// getJWTSecret возвращает JWT секрет из переменных окружения
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback для разработки
		return []byte("fallback-secret-key-change-in-production")
	}
	return []byte(secret)
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email, role string) (string, error) {
	accessExp, _ := GetTokenExpiration()
	expirationTime := time.Now().Add(accessExp)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   email,
			Issuer:    "auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func GenerateRefreshToken() (string, error) {
	// Генерируем 32 случайных байта
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
