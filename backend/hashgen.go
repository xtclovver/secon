package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := []string{"admin", "manager", "user"} // Ваши пароли

	for _, pass := range passwords {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Ошибка хеширования пароля %s: %v", pass, err)
		}
		fmt.Printf("Пароль: %s\nХеш:    %s\n\n", pass, string(hashedPassword))
	}
}