package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plaintext password with bcrypt
func HashPassword(password string) (string, error) {
b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
return "", err
}
return string(b), nil
}

// CheckPassword compares plaintext password with bcrypt hash
func CheckPassword(hash, password string) bool {
return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}