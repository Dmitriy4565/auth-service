package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"time"
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
	fmt.Printf("\n🎯 ОТПРАВКА 2FA КОДА ЧЕРЕЗ YANDEX: %s -> %s\n", code, email)

	// ЗАПУСКАЕМ В ОТДЕЛЬНОЙ ГОРУТИНЕ - НЕ БЛОКИРУЕМ ОСНОВНОЙ ПОТОК
	go s.send2FACodeAsync(email, code)

	return nil // сразу возвращаем успех
}

func (s *EmailService) SendResetPasswordEmail(email, resetLink string) error {
	fmt.Printf("\n🔐 ОТПРАВКА ССЫЛКИ СБРОСА ЧЕРЕЗ YANDEX: %s -> %s\n", email, resetLink)

	// ТОЖЕ В ОТДЕЛЬНОЙ ГОРУТИНЕ
	go s.sendResetPasswordAsync(email, resetLink)

	return nil
}

// Асинхронные методы (работают в фоне)
func (s *EmailService) send2FACodeAsync(email, code string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Паника при отправке 2FA: %v\n", r)
		}
	}()

	start := time.Now()
	fmt.Printf("📧 [YANDEX] Отправляем 2FA код на %s\n", email)

	// HTML содержимое
	htmlContent := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1890ff;">Ростелеком Проекты</h2>
			<h3>Ваш код подтверждения</h3>
			<div style="font-size: 32px; font-weight: bold; color: #1890ff; text-align: center; margin: 20px 0; padding: 10px; background: #f5f5f5;">
				%s
			</div>
			<p><strong>Код действителен 10 минут</strong></p>
			<p style="color: #666; font-size: 12px; margin-top: 20px;">
				Если вы не запрашивали этот код, проигнорируйте это письмо.
			</p>
		</div>`, code)

	err := s.sendEmailWithTimeout(
		email,
		"Код двухфакторной аутентификации - Ростелеком Проекты",
		htmlContent,
	)

	if err != nil {
		fmt.Printf("❌ [YANDEX] Ошибка отправки 2FA на %s: %v\n", email, err)
	} else {
		fmt.Printf("✅ [YANDEX] Письмо с кодом %s отправлено на %s за %v\n",
			code, email, time.Since(start))
	}
}

func (s *EmailService) sendResetPasswordAsync(email, resetLink string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Паника при отправке reset: %v\n", r)
		}
	}()

	start := time.Now()
	fmt.Printf("🔐 [YANDEX] Отправляем reset ссылку на %s\n", email)

	// HTML содержимое
	htmlContent := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd;">
        <h2 style="color: #1890ff;">Ростелеком Проекты</h2>
        <h3>Сброс пароля</h3>
        <p>Для сброса пароля перейдите по ссылке ниже:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #1890ff; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-size: 16px; display: inline-block;">
                Сбросить пароль
            </a>
        </div>
        <p><strong>Ссылка действительна 1 час.</strong></p>
        <p>Если вы не запрашивали сброс пароля, проигнорируйте это письмо.</p>
        <hr>
        <p style="color: #666; font-size: 12px;">Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
    </div>
</body>
</html>`, resetLink)

	err := s.sendEmailWithTimeout(
		email,
		"Сброс пароля - Ростелеком Проекты",
		htmlContent,
	)

	if err != nil {
		fmt.Printf("❌ [YANDEX] Ошибка отправки reset на %s: %v\n", email, err)
	} else {
		fmt.Printf("✅ [YANDEX] Ссылка сброса отправлена на %s за %v\n",
			email, time.Since(start))
	}
}

// Отправка с таймаутом
func (s *EmailService) sendEmailWithTimeout(to, subject, html string) error {
	// Создаем канал для результата
	result := make(chan error, 1)

	// Запускаем отправку в отдельной горутине
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("panic: %v", r)
			}
		}()

		err := s.sendEmailSMTP(to, subject, html)
		result <- err
	}()

	// Ждем с таймаутом 15 секунд
	select {
	case err := <-result:
		return err
	case <-time.After(15 * time.Second):
		return fmt.Errorf("таймаут отправки письма")
	}
}

// Базовая отправка через Яндекс SMTP
func (s *EmailService) sendEmailSMTP(to, subject, html string) error {
	// Проверяем настройки
	if s.host == "" || s.username == "" || s.password == "" {
		return fmt.Errorf("SMTP настройки не заполнены")
	}

	// Аутентификация
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Формируем сообщение
	emailSubject := "Subject: " + subject + "\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	msg := []byte(emailSubject + mime + html)

	fmt.Printf("📤 [YANDEX] Подключаемся к %s:%s...\n", s.host, s.port)

	// Яндекс требует TLS, поэтому используем продвинутую отправку
	err := s.sendWithTLS(to, msg, auth)
	if err != nil {
		return fmt.Errorf("ошибка отправки через Яндекс SMTP: %v", err)
	}

	fmt.Printf("✅ [YANDEX] Письмо успешно отправлено через Яндекс SMTP\n")
	return nil
}

func (s *EmailService) sendWithTLS(to string, msg []byte, auth smtp.Auth) error {
	// Подключаемся к SMTP серверу
	client, err := smtp.Dial(s.host + ":" + s.port)
	if err != nil {
		return err
	}
	defer client.Close()

	// STARTTLS (Яндекс требует шифрование)
	if err = client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
		return err
	}

	// Аутентификация
	if err = client.Auth(auth); err != nil {
		return err
	}

	// Устанавливаем отправителя и получателя
	if err = client.Mail(s.from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}

	// Отправляем данные
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
