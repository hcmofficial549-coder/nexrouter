# 🚀 nexrouter

**High-performance HTTP router for Go with zero allocation on hot path**

## ✨ Features

- 🌳 Radix Tree Routing — O(k) lookup, zero allocation
- ⚙️ Middleware Chain — Composable with Next() support
- 📁 Route Groups — Shared prefix & middleware
- 🛡️ Panic Recovery — Automatic recovery from panics
- 🌐 CORS — Production-ready CORS middleware
- 🚦 Rate Limiting — Token bucket algorithm
- 🔐 JWT Authentication — Built-in JWT middleware
- 📝 Request Logger — Structured logging

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

