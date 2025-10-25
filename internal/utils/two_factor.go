package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"time"

	"github.com/pquerna/otp/totp"
)

func GenerateTwoFactorSecret() (string, error) {
	randomBytes := make([]byte, 20)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes), nil
}

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

func ValidateTwoFactorCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

func GenerateTwoFactorCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

func ValidateCustomTwoFactorCode(inputCode, storedCode string, expiresAt time.Time) bool {
	if time.Now().After(expiresAt) {
		return false
	}
	return inputCode == storedCode
}
