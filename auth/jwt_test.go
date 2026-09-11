package auth

import (
"testing"
"time"
)

func TestGenerateAndValidate(t *testing.T) {
secret := "test-secret"
token, err := GenerateToken("1", "a@b.com", "Alice", secret, time.Hour)
if err != nil {
t.Fatal(err)
}
claims, err := ValidateToken(token, secret)
if err != nil {
t.Fatal(err)
}
if claims.Email != "a@b.com" || claims.Name != "Alice" || claims.Sub != "1" {
t.Errorf("claims mismatch: %+v", claims)
}
}

func TestInvalidSignature(t *testing.T) {
token, _ := GenerateToken("1", "a@b.com", "A", "secret-1", time.Hour)
if _, err := ValidateToken(token, "secret-2"); err == nil {
t.Error("expected error for wrong secret")
}
}

func TestExpiredToken(t *testing.T) {
token, _ := GenerateToken("1", "a@b.com", "A", "s", -time.Hour)
if _, err := ValidateToken(token, "s"); err == nil {
t.Error("expected error for expired token")
}
}

func TestPasswordHash(t *testing.T) {
hash, err := HashPassword("password123")
if err != nil {
t.Fatal(err)
}
if hash == "password123" {
t.Error("password stored in plaintext!")
}
if !CheckPassword(hash, "password123") {
t.Error("correct password should match")
}
if CheckPassword(hash, "wrongpass") {
t.Error("wrong password should not match")
}
}