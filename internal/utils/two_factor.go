package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"time"

	"github.com/pquerna/otp/totp"
)

// GenerateTwoFactorSecret создает случайный секрет для TOTP (Time-based One-Time Password)
func GenerateTwoFactorSecret() (string, error) {
	// Генерируем 20 случайных байт
	randomBytes := make([]byte, 20)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Кодируем в base32 без padding для совместимости с аутентификаторами
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes), nil
}

// GenerateQRCode создает QR код в формате URL для приложения аутентификатора
func GenerateQRCode(secret, email, issuer string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		Secret:      []byte(secret),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}
	return key.URL(), nil
}

// ValidateTwoFactorCode проверяет TOTP код по секрету
func ValidateTwoFactorCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateTwoFactorCode создает 6-значный код для SMS/Email аутентификации
func GenerateTwoFactorCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)

	for i := range code {
		// Генерируем случайное число от 0 до 9
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

// ValidateCustomTwoFactorCode проверяет временный код (SMS/Email)
func ValidateCustomTwoFactorCode(inputCode, storedCode string, expiresAt time.Time) bool {
	// Проверяем не истекло ли время действия кода
	if time.Now().After(expiresAt) {
		return false
	}
	return inputCode == storedCode
}
