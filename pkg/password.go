package pkg

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordManager struct{}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{}
}

// HashPassword hashes a plain-text password
func (p *PasswordManager) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// CheckPassword verifies a password against its hashed version
func (p *PasswordManager) CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
