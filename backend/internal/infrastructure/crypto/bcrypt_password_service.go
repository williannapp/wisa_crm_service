package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordService implements domain.PasswordService using bcrypt.
type BcryptPasswordService struct{}

// NewBcryptPasswordService creates a new BcryptPasswordService.
func NewBcryptPasswordService() *BcryptPasswordService {
	return &BcryptPasswordService{}
}

// Compare returns true if plain matches the bcrypt hashed password.
func (s *BcryptPasswordService) Compare(plain, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}
