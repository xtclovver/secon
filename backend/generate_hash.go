package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
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
