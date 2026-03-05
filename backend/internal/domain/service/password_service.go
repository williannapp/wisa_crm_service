package service

// PasswordService defines the port for password verification.
// Used for comparing plain password with bcrypt hash.
// Per ADR-010: use case must always call Compare (with dummy hash if user not found)
// to prevent timing attacks.
type PasswordService interface {
	Compare(plain, hashed string) bool
}
