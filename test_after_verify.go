package main

import (
	"auth-service/internal/service"
	"fmt"
	"time"
)

func main() {
	fmt.Println("🕐 ТЕСТ ПОСЛЕ ВЕРИФИКАЦИИ ДОМЕНА")
	fmt.Println("================================")

	emailService := service.NewEmailService()

	emails := []string{
		"a.d1mon@yandex.ru",
		"d.prihodin@yandex.ru",
		"prihodin816@gmail.com",
	}

	for i, email := range emails {
		fmt.Printf("\n%d. 📧 Отправляем на %s\n", i+1, email)
		err := emailService.Send2FACode(email, "999888")
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n", err)
		} else {
			fmt.Println("✅ Запрос отправлен!")
		}
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n🎉 ВСЕ ЗАПРОСЫ ОТПРАВЛЕНЫ!")
	fmt.Println("Проверяйте почтовые ящики!")
}
