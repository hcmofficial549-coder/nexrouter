package middleware

import "github.com/nexrouter/nexrouter/core"

// Security adds standard security headers to every response
func Security() core.HandlerFunc {
return func(c *core.Context) {
c.SetHeader("X-Content-Type-Options", "nosniff")
c.SetHeader("X-Frame-Options", "SAMEORIGIN")
c.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
c.SetHeader("X-XSS-Protection", "1; mode=block")
c.Next()
}
}