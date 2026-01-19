package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func main() {
	passwords := map[string]string{
		"admin": "Admin@123456",
		"user1": "User@123456",
		"user2": "User@123456",
	}

	for username, password := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", username, string(hash))
	}
}
