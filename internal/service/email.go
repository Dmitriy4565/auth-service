package service

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *EmailService) Send2FACode(email, code string) error {
	// Если SMTP не настроен, выводим в консоль (демо-режим)
	if s.host == "" || s.username == "" || s.password == "" {
		fmt.Printf("🎯 2FA код для %s: %s (SMTP не настроен)\n", email, code)
		return nil
	}

	// Настройка SMTP
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	port, _ := strconv.Atoi(s.port)
	if port == 0 {
		port = 587
	}

	// Текст письма
	subject := "Subject: Код двухфакторной аутентификации - Ростелеком Проекты\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd;">
        <h2 style="color: #1890ff;">Ростелеком Проекты</h2>
        <h3>Код двухфакторной аутентификации</h3>
        <p>Ваш код для входа в систему:</p>
        <div style="font-size: 32px; font-weight: bold; color: #1890ff; text-align: center; margin: 20px 0;">
            %s
        </div>
        <p><strong>Код действителен 10 минут.</strong></p>
        <p>Если вы не запрашивали этот код, проигнорируйте это письмо.</p>
        <hr>
        <p style="color: #666; font-size: 12px;">Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
    </div>
</body>
</html>`, code)

	msg := []byte(subject + mime + body)

	// Отправка почты
	err := smtp.SendMail(s.host+":"+strconv.Itoa(port), auth, s.from, []string{email}, msg)
	if err != nil {
		fmt.Printf("❌ Ошибка отправки почты: %v\n", err)
		return err
	}

	fmt.Printf("✅ Код 2FA отправлен на %s\n", email)
	return nil
}
