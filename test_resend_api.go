package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")

	fmt.Printf("🔑 API Key: %s...\n", getFirstChars(apiKey, 10))
	fmt.Printf("📧 From Email: %s\n", fromEmail)

	if apiKey == "" {
		fmt.Println("❌ RESEND_API_KEY не установлен")
		return
	}

	// Prepare email
	emailReq := map[string]interface{}{
		"from":    "Auth Service <" + fromEmail + ">",
		"to":      []string{"prihodin816@gmail.com"},
		"subject": "Тест Resend API - Ростелеком Проекты",
		"html":    `<div style="font-family: Arial, sans-serif;"><h2 style="color: #1890ff;">Ростелеком Проекты</h2><h3>Тест Resend API</h3><p><strong>Работает нахуй через прямые запросы! 🚀</strong></p><p>Если ты это видишь - всё работает!</p></div>`,
		"text":    "Ростелеком Проекты\nТест Resend API\nРаботает нахуй через прямые запросы! 🚀\nЕсли ты это видишь - всё работает!",
	}

	jsonData, _ := json.Marshal(emailReq)

	// Send request
	req, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	fmt.Println("📤 Отправляем тестовое письмо через Resend API...")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("📊 Status: %d\n", resp.StatusCode)
	fmt.Printf("📨 Response: %s\n", string(body))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("✅ Успешно! Проверяй почту prihodin816@gmail.com!")
	} else {
		fmt.Println("❌ Ошибка отправки")
	}
}

func getFirstChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
