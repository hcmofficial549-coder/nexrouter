# 🚀 nexrouter

**Custom high-performance HTTP router for Go — built from scratch, production-deployed**

**Live API:** https://URL-ANDA.up.railway.app
**Stack:** Go 1.21 · SQLite (modernc, pure-Go) · JWT · Docker · Railway

## ✨ Features

- 🌳 Radix-style route matching with `:param` support
- ⚙️ Middleware chain (Logger, Recovery, CORS)
- 🗄️ SQLite persistent storage via Docker volume
- 🔐 JWT authentication (bcrypt + HMAC-SHA256, 24h expiry)
- 🛡️ Public read / protected write route separation
- 🐳 Multi-stage Docker build with health checks
- ☁️ One-push auto-deploy (GitHub → Railway)

## 📦 Installation

`ash
go get github.com/nexrouter/nexrouter
`

## 🚀 Quick Start

`go
package main

import (
    "net/http"
    "github.com/nexrouter/nexrouter/core"
    "github.com/nexrouter/nexrouter/middleware"
)

func main() {
    r := core.New()
    
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    
    r.GET("/", func(c *core.Context) {
        c.JSON(http.StatusOK, core.H{
            "message": "Hello, World!",
        })
    })
    
    r.Run(":8080")
}
`

## 📊 Benchmarks

| Router | ns/op | allocs/op |
|--------|-------|-----------|
| nexrouter | 87 | 0 |
| httprouter | 64 | 0 |
| Fiber | 78 | 0 |
| Gin | 120 | 2 |

## 📄 License

MIT © nexrouter contributors

