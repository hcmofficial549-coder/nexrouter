# nexrouter

Custom high-performance HTTP router for Go - built from scratch, production-deployed.

[![CI](https://github.com/hcmofficial549-coder/nexrouter/actions/workflows/ci.yml/badge.svg)](https://github.com/hcmofficial549-coder/nexrouter/actions/workflows/ci.yml)

## Live Demo

- Dashboard: https://nexrouter.up.railway.app/dashboard
- API Docs: https://nexrouter.up.railway.app/docs
- API Docs: https://nexrouter.up.railway.app/docs
- API Health: https://nexrouter.up.railway.app/health
- Demo login: john@example.com / password123
- Live Monitor: https://nexrouter.up.railway.app/monitor
- Chat App: https://nexchat-production.up.railway.app

## Features

- Radix-style route matching with :param support (built from scratch, no Gin/Echo)
- Middleware chain: Logger, Recovery, CORS, Security Headers
- Rate limiting: token bucket algorithm, 60 requests/min per IP
- SQLite persistent storage (modernc pure-Go driver) via Docker volume
- JWT authentication: bcrypt password hashing + HMAC-SHA256 tokens, 24h expiry
- Public read / protected write route separation
- Embedded dashboard UI (go:embed) served at /dashboard
- Multi-stage Docker build with health checks
- Unit tests with race detector + GitHub Actions CI
- Auto-deploy: git push triggers Railway deployment

## Quick Start