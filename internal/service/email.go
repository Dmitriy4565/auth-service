package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
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
	fmt.Printf("🎯 ОТПРАВКА 2FA КОДА:\n")
	fmt.Printf("📧 Кому: %s\n", email)
	fmt.Printf("🔐 Код: %s\n", code)
	fmt.Printf("⚙️ SMTP: %s:%s\n", s.host, s.port)
	fmt.Printf("👤 Auth: %s\n", s.username)

	// Проверяем настройки
	if s.host == "" || s.username == "" || s.password == "" {
		fmt.Printf("❌ SMTP настройки не заполнены!\n")
		fmt.Printf("   HOST: '%s'\n", s.host)
		fmt.Printf("   USER: '%s'\n", s.username)
		fmt.Printf("   PASS: '%s'\n", "***") // не показываем пароль
		return fmt.Errorf("SMTP настройки не заполнены")
	}

	// Аутентификация
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	port, err := strconv.Atoi(s.port)
	if err != nil || port == 0 {
		port = 587
		fmt.Printf("⚙️ Используем порт по умолчанию: %d\n", port)
	}

	// Формируем сообщение
	subject := "Subject: Код двухфакторной аутентификации\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1890ff;">Ростелеком Проекты</h2>
			<h3>Ваш код подтверждения</h3>
			<div style="font-size: 32px; font-weight: bold; color: #1890ff; text-align: center; margin: 20px 0; padding: 10px; background: #f5f5f5;">
				%s
			</div>
			<p><strong>Код действителен 10 минут</strong></p>
		</div>`, code)

	msg := []byte(subject + mime + body)

	// Пытаемся отправить
	fmt.Printf("📤 Пытаемся отправить письмо...\n")

	// Пробуем разные способы отправки

	// Способ 1: Обычная отправка
	err = smtp.SendMail(s.host+":"+strconv.Itoa(port), auth, s.from, []string{email}, msg)
	if err != nil {
		fmt.Printf("❌ Способ 1 не сработал: %v\n", err)

		// Способ 2: С TLS
		fmt.Printf("🔄 Пробуем с TLS...\n")
		err = s.sendWithTLS(email, msg, auth, port)
		if err != nil {
			fmt.Printf("❌ Способ 2 тоже не сработал: %v\n", err)
			return fmt.Errorf("не удалось отправить email: %v", err)
		}
	}

	fmt.Printf("✅ Письмо успешно отправлено на %s\n", email)
	fmt.Printf("📨 Код: %s\n", code)
	return nil
}

// sendWithTLS - отправка с явным TLS
func (s *EmailService) sendWithTLS(to string, msg []byte, auth smtp.Auth, port int) error {
	// Подключаемся к SMTP серверу
	client, err := smtp.Dial(s.host + ":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer client.Close()

	// STARTTLS
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

// SendResetPasswordEmail отправляет email для сброса пароля
func (s *EmailService) SendResetPasswordEmail(email, resetLink string) error {
	fmt.Printf("🔐 ОТПРАВКА ССЫЛКИ СБРОСА ПАРОЛЯ:\n")
	fmt.Printf("📧 Кому: %s\n", email)
	fmt.Printf("🔗 Ссылка: %s\n", resetLink)
	fmt.Printf("⚙️ SMTP: %s:%s\n", s.host, s.port)
	fmt.Printf("👤 Auth: %s\n", s.username)

	// Проверяем настройки
	if s.host == "" || s.username == "" || s.password == "" {
		fmt.Printf("❌ SMTP настройки не заполнены!\n")
		return fmt.Errorf("SMTP настройки не заполнены")
	}

	// Аутентификация
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	port, err := strconv.Atoi(s.port)
	if err != nil || port == 0 {
		port = 587
		fmt.Printf("⚙️ Используем порт по умолчанию: %d\n", port)
	}

	// Текст письма
	subject := "Subject: Сброс пароля - Ростелеком Проекты\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
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

	msg := []byte(subject + mime + body)

	// Отправка почты
	fmt.Printf("📤 Пытаемся отправить письмо...\n")
	err = smtp.SendMail(s.host+":"+strconv.Itoa(port), auth, s.from, []string{email}, msg)
	if err != nil {
		fmt.Printf("❌ Ошибка отправки почты для сброса пароля: %v\n", err)
		return fmt.Errorf("ошибка отправки email: %v", err)
	}

	fmt.Printf("✅ Ссылка сброса пароля отправлена на %s\n", email)
	fmt.Printf("🔑 ТОКЕН ДЛЯ ТЕСТА: %s\n", strings.Split(resetLink, "/reset-password/")[1])
	return nil
}
