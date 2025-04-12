package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := []string{"admin", "manager", "user", "pass1", "pass2", "pass3", "pass4"} // Ваши пароли

	for _, pass := range passwords {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Ошибка хеширования пароля %s: %v", pass, err)
		}
		fmt.Printf("Пароль: %s\nХеш:    %s\n\n", pass, string(hashedPassword))
	}
}
