package database

import (
	"database/sql"
	"fmt"
	"log"
	"time" // Добавляем импорт пакета time

	_ "github.com/go-sql-driver/mysql" // Импортируем драйвер MySQL

	"vacation-scheduler/internal/config"
)

// NewConnection создает и возвращает новое подключение к базе данных
func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	log.Println("Попытка подключения к базе данных...") // Логирование

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		log.Printf("Ошибка при открытии соединения с БД: %v\n", err)
		return nil, fmt.Errorf("ошибка открытия соединения с БД: %w", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		log.Printf("Ошибка при проверке соединения с БД (Ping): %v\n", err)
		db.Close() // Закрываем соединение, если пинг не прошел
		return nil, fmt.Errorf("ошибка проверки соединения с БД: %w", err)
	}

	log.Println("Успешное подключение к базе данных!")
	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)                 // Максимальное количество открытых соединений
	db.SetMaxIdleConns(25)                 // Максимальное количество простаивающих соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения

	return db, nil
}
