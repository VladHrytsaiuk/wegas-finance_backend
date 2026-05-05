package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// Ключ шифрування. В продакшені бери з ENV!
func getSecretKey() []byte {
	key := os.Getenv("APP_SECRET_KEY")
	if len(key) != 32 {
		return []byte("12345678901234567890123456789012") // 32 bytes mock key
	}
	return []byte(key)
}

func Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(cryptoText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}