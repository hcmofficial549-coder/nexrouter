package auth

import (
"net/http"
"strings"

"github.com/nexrouter/nexrouter/core"
)

// JWT returns middleware that validates Bearer tokens
func JWT(secret string) core.HandlerFunc {
return func(c *core.Context) {
h := c.GetHeader("Authorization")
if h == "" || !strings.HasPrefix(h, "Bearer ") {
c.JSON(http.StatusUnauthorized, core.H{"error": "missing Authorization Bearer token"})
c.Abort()
return
}
token := strings.TrimPrefix(h, "Bearer ")
claims, err := ValidateToken(token, secret)
if err != nil {
c.JSON(http.StatusUnauthorized, core.H{"error": err.Error()})
c.Abort()
return
}
c.Set("claims", claims)
c.Next()
}
}

// GetClaims retrieves validated claims from context
func GetClaims(c *core.Context) *Claims {
v, ok := c.Get("claims")
if !ok {
return nil
}
claims, _ := v.(*Claims)
return claims
}