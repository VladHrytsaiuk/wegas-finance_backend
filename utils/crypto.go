package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword - перетворює "1234" на "$2a$10$..."
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword - перевіряє, чи підходить пароль до хешу
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}