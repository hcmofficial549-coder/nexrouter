# 🌳 nexrouter

Custom high-performance HTTP router for Go — built from scratch, production-deployed.

**Live API:** https://nexrouter.up.railway.app
**Stack:** Go 1.21 · SQLite (modernc, pure-Go) · JWT · Docker · Railway

## ✨ Features
- 🌳 Radix-style route matching with `:param` support
- ⚙️ Middleware chain (Logger, Recovery, CORS)
- 🗄️ SQLite persistent storage via Docker volume
- 🔐 JWT authentication (bcrypt + HMAC-SHA256, 24h expiry)
- 🛡️ Public read / protected write route separation
- 🐳 Multi-stage Docker build with health checks
- ☁️ One-push auto-deploy (GitHub → Railway)

## 🚀 Quick Start

