package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type EmailService struct {
	apiKey string
	from   string
	name   string
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	fromName := os.Getenv("RESEND_FROM_NAME")

	if fromName == "" {
		fromName = "Auth Service"
	}

	fmt.Printf("🔧 Инициализация Resend: %s...\n", getFirstChars(apiKey, 10))

	return &EmailService{
		apiKey: apiKey,
		from:   fromEmail,
		name:   fromName,
	}
}

func (s *EmailService) Send2FACode(email, code string) error {
	fmt.Printf("\n🎯 ОТПРАВКА 2FA КОДА ЧЕРЕЗ RESEND: %s -> %s\n", code, email)

	// Асинхронная отправка
	go s.send2FACodeAsync(email, code)
	return nil
}

func (s *EmailService) SendResetPasswordEmail(email, resetLink string) error {
	fmt.Printf("\n🔐 ОТПРАВКА ССЫЛКИ СБРОСА ЧЕРЕЗ RESEND: %s -> %s\n", email, resetLink)

	// Асинхронная отправка
	go s.sendResetPasswordAsync(email, resetLink)
	return nil
}

func (s *EmailService) send2FACodeAsync(email, code string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Паника при отправке 2FA: %v\n", r)
		}
	}()

	start := time.Now()
	fmt.Printf("📧 [RESEND] Отправляем 2FA код на %s\n", email)

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

	plainTextContent := fmt.Sprintf(
		"Ростелеком Проекты\nВаш код подтверждения: %s\nКод действителен 10 минут",
		code,
	)

	err := s.sendEmailWithTimeout(
		email,
		"Код двухфакторной аутентификации - Ростелеком Проекты",
		htmlContent,
		plainTextContent,
	)

	if err != nil {
		fmt.Printf("❌ [RESEND] Ошибка отправки 2FA на %s: %v\n", email, err)
	} else {
		fmt.Printf("✅ [RESEND] Письмо с кодом %s отправлено на %s за %v\n",
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
	fmt.Printf("🔐 [RESEND] Отправляем reset ссылку на %s\n", email)

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

	plainTextContent := fmt.Sprintf(
		"Ростелеком Проекты\nСброс пароля\nДля сброса пароля перейдите по ссылке: %s\nСсылка действительна 1 час.",
		resetLink,
	)

	err := s.sendEmailWithTimeout(
		email,
		"Сброс пароля - Ростелеком Проекты",
		htmlContent,
		plainTextContent,
	)

	if err != nil {
		fmt.Printf("❌ [RESEND] Ошибка отправки reset на %s: %v\n", email, err)
	} else {
		fmt.Printf("✅ [RESEND] Ссылка сброса отправлена на %s за %v\n",
			email, time.Since(start))
	}
}

func (s *EmailService) sendEmailWithTimeout(to, subject, html, text string) error {
	// Создаем канал для результата
	result := make(chan error, 1)

	// Запускаем отправку в отдельной горутине
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("panic: %v", r)
			}
		}()

		err := s.sendEmailResend(to, subject, html, text)
		result <- err
	}()

	// Ждем с таймаутом 10 секунд
	select {
	case err := <-result:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("таймаут отправки письма")
	}
}

// Resend API структуры
type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
	Text    string   `json:"text"`
}

type ResendEmailResponse struct {
	Id string `json:"id"`
}

func (s *EmailService) sendEmailResend(to, subject, html, text string) error {
	// Проверяем настройки
	if s.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY не установлен")
	}

	if s.from == "" {
		return fmt.Errorf("RESEND_FROM_EMAIL не установлен")
	}

	// Prepare request
	emailReq := ResendEmailRequest{
		From:    s.name + " <" + s.from + ">",
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	fmt.Printf("📤 [RESEND] Отправляем запрос к API...\n")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	fmt.Printf("📊 [RESEND] Status Code: %d\n", resp.StatusCode)

	// Check status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var emailResp ResendEmailResponse
		if err := json.Unmarshal(body, &emailResp); err == nil {
			fmt.Printf("✅ [RESEND] ID письма: %s\n", emailResp.Id)
		} else {
			fmt.Printf("✅ [RESEND] Письмо отправлено (не удалось распарсить ID)\n")
		}
		return nil
	} else {
		fmt.Printf("❌ [RESEND] Response Body: %s\n", string(body))
		return fmt.Errorf("Resend error %d: %s", resp.StatusCode, string(body))
	}
}

func getFirstChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
