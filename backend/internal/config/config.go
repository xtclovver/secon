package config

import (
	// В реальном приложении здесь будут импорты для чтения конфигурации (например, viper)
	"errors"
)

// Config - структура для хранения конфигурации приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig - конфигурация сервера
type ServerConfig struct {
	Port string
}

// DatabaseConfig - конфигурация базы данных
type DatabaseConfig struct {
	DSN string // Data Source Name (e.g., "user:password@tcp(localhost:3306)/vacation_scheduler?parseTime=true")
}

// JWTConfig - конфигурация JWT
type JWTConfig struct {
	Secret string // Секретный ключ для подписи токенов
}

// Load - функция для загрузки конфигурации (заглушка)
// В реальном приложении здесь будет логика чтения из файла (e.g., config.yaml) или переменных окружения
func Load() (*Config, error) {
	// Заглушка с дефолтными значениями
	cfg := &Config{
		Server: ServerConfig{
			Port: ":8080", // Порт по умолчанию
		},
		Database: DatabaseConfig{
			// ВАЖНО: Замените на ваш реальный DSN для MySQL
			DSN: "root:12341234@tcp(localhost:3306)/vacation_scheduler?parseTime=true", 
		},
		JWT: JWTConfig{
			Secret: "your_very_secret_jwt_key", // ВАЖНО: Замените на ваш секретный ключ! Лучше брать из env.
		},
	}

	// Простая валидация (пример)
	if cfg.Database.DSN == "user:password@tcp(localhost:3306)/vacation_scheduler?parseTime=true" {
		// Можно выводить предупреждение, но не блокировать запуск для простоты
		// log.Println("ПРЕДУПРЕЖДЕНИЕ: Используется DSN базы данных по умолчанию. Укажите реальные данные.")
	}
	if cfg.JWT.Secret == "your_very_secret_jwt_key" {
		// log.Println("ПРЕДУПРЕЖДЕНИЕ: Используется секретный ключ JWT по умолчанию. Замените его на надежный ключ.")
		// return nil, errors.New("необходимо установить надежный секретный ключ JWT") // Можно сделать обязательным
	}
	if cfg.Server.Port == "" {
		return nil, errors.New("необходимо указать порт сервера")
	}


	return cfg, nil
}
