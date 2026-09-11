package middleware

import (
"log"
"net/http"
"runtime"
"time"

"github.com/nexrouter/nexrouter/core"
)

// Logger logs each request
func Logger() core.HandlerFunc {
return func(c *core.Context) {
start := time.Now()
path := c.Request.URL.Path
c.Next()
log.Printf("[nexrouter] %3d | %13v | %15s | %-7s %s",
c.StatusCode(), time.Since(start), c.ClientIP(), c.Request.Method, path)
}
}

// Recovery recovers from panics
func Recovery() core.HandlerFunc {
return func(c *core.Context) {
defer func() {
if err := recover(); err != nil {
stack := make([]byte, 4096)
n := runtime.Stack(stack, false)
log.Printf("[nexrouter] PANIC: %v\n%s", err, stack[:n])
if !c.Written() {
c.JSON(http.StatusInternalServerError, core.H{"error": "internal server error"})
}
c.Abort()
}
}()
c.Next()
}
}

// CORS adds permissive CORS headers
func CORS() core.HandlerFunc {
return func(c *core.Context) {
c.SetHeader("Access-Control-Allow-Origin", "*")
c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
if c.Request.Method == http.MethodOptions {
c.AbortWithStatus(http.StatusNoContent)
return
}
c.Next()
}
}