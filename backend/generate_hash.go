package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// generateHash генерирует хеш пароля из аргумента командной строки.
// Примечание: Эта функция была переименована из 'main', чтобы избежать конфликта
// с основной функцией main в пакете. Для использования этого скрипта,
// вы можете временно переименовать его обратно в 'main' или вызвать
// эту функцию из другого места.
func generateHash() {
	// Пароль берется из первого аргумента командной строки
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run generate_hash.go <password>")
		os.Exit(1)
	}
	password := os.Args[1] // Пароль, который нужно хешировать

	// Генерируем хеш пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
		os.Exit(1)
	}

	// Выводим хеш в стандартный вывод
	fmt.Println(string(hashedPassword))
}
