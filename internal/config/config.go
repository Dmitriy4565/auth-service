package config

import (
    "os"
    "strconv"
)

// Config хранит все конфигурационные параметры приложения
type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    Port       string
    JWTSecret  string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
    return &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", "password"),
        DBName:     getEnv("DB_NAME", "auth_service"),
        Port:       getEnv("PORT", "8080"),
        JWTSecret:  getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
    }
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// getEnvAsInt возвращает переменную окружения как integer
func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}