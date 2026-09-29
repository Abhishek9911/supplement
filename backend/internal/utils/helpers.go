package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateOrderNumber() string {
	return fmt.Sprintf("SUP-%d-%s", time.Now().Unix(), randomHex(6))
}

func randomHex(length int) string {
	buf := make([]byte, length)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
