package auth

import (
"crypto/hmac"
"crypto/sha256"
"encoding/base64"
"encoding/json"
"errors"
"strings"
"time"
)

// Claims represents the JWT payload
type Claims struct {
Sub   string `json:"sub"`
Email string `json:"email"`
Name  string `json:"name"`
Iat   int64  `json:"iat"`
Exp   int64  `json:"exp"`
}

func b64(b []byte) string {
return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateToken creates a signed JWT (HS256)
func GenerateToken(sub, email, name, secret string, ttl time.Duration) (string, error) {
hb, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
now := time.Now()
cb, err := json.Marshal(Claims{Sub: sub, Email: email, Name: name, Iat: now.Unix(), Exp: now.Add(ttl).Unix()})
if err != nil {
return "", err
}
signingInput := b64(hb) + "." + b64(cb)
mac := hmac.New(sha256.New, []byte(secret))
mac.Write([]byte(signingInput))
return signingInput + "." + b64(mac.Sum(nil)), nil
}

// ValidateToken verifies signature and expiration
func ValidateToken(tokenString, secret string) (*Claims, error) {
parts := strings.Split(tokenString, ".")
if len(parts) != 3 {
return nil, errors.New("invalid token format")
}
signingInput := parts[0] + "." + parts[1]
mac := hmac.New(sha256.New, []byte(secret))
mac.Write([]byte(signingInput))
if !hmac.Equal([]byte(parts[2]), []byte(b64(mac.Sum(nil)))) {
return nil, errors.New("invalid token signature")
}
payload, err := base64.RawURLEncoding.DecodeString(parts[1])
if err != nil {
return nil, errors.New("invalid token payload")
}
var claims Claims
if err := json.Unmarshal(payload, &claims); err != nil {
return nil, errors.New("invalid token claims")
}
if claims.Exp < time.Now().Unix() {
return nil, errors.New("token expired")
}
return &claims, nil
}