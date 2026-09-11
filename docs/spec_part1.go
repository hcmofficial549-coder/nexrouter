package docs

const specPart1 = `
{
  "openapi": "3.0.3",
  "info": {
    "title": "nexrouter API",
    "description": "Custom high-performance HTTP router for Go - built from scratch. REST API with JWT authentication, SQLite persistence, caching layer, and realtime SSE events. Protected endpoints require: Authorization Bearer token (get one from /api/v1/auth/login).",
    "version": "1.7.0",
    "contact": {
      "name": "hcmofficial549-coder",
      "url": "https://github.com/hcmofficial549-coder/nexrouter"
    }
  },
  "servers": [
    { "url": "https://nexrouter.up.railway.app", "description": "Production (Railway)" },
    { "url": "http://localhost:8080", "description": "Local development" }
  ],
  "tags": [
    { "name": "auth", "description": "Register, login, profile" },
    { "name": "users", "description": "User management" },
    { "name": "products", "description": "Product CRUD (writes protected)" },
    { "name": "system", "description": "Health, stats, cache" }
  ],
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "Paste the token from /api/v1/auth/login response"
      }
    },
    "schemas": {
      "User": {
        "type": "object",
        "properties": {
          "id": { "type": "integer", "example": 1 },
          "name": { "type": "string", "example": "John Doe" },
          "email": { "type": "string", "example": "john@example.com" }
        }
      },
      "Product": {
        "type": "object",
        "properties": {
          "id": { "type": "integer", "example": 1 },
          "name": { "type": "string", "example": "Laptop" },
          "price": { "type": "number", "format": "double", "example": 999.99 },
          "stock": { "type": "integer", "example": 10 },
          "category": { "type": "string", "example": "electronics" }
        }
      },
      "Stats": {
        "type": "object",
        "properties": {
          "users": { "type": "integer", "example": 2 },
          "products": { "type": "integer", "example": 3 },
          "total_stock": { "type": "integer", "example": 68 },
          "inventory_value": { "type": "number", "example": 13093.34 }
        }
      },
      "Error": {
        "type": "object",
        "properties": {
          "error": { "type": "string", "example": "not found" }
        }
      },
      "RegisterRequest": {
        "type": "object",
        "required": ["name", "email", "password"],
        "properties": {
          "name": { "type": "string", "example": "Siti Rahayu" },
          "email": { "type": "string", "example": "siti@test.com" },
          "password": { "type": "string", "format": "password", "example": "rahasia123", "description": "min 6 characters" }
        }
      },
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": { "type": "string", "example": "john@example.com" },
          "password": { "type": "string", "format": "password", "example": "password123" }
        }
      },
      "TokenResponse": {
        "type": "object",
        "properties": {
          "success": { "type": "boolean", "example": true },
          "token": { "type": "string", "description": "JWT - send as header Authorization: Bearer TOKEN" },
          "expires_in": { "type": "string", "example": "24h" },
          "user": { "$ref": "#/components/schemas/User" }
        }
      },
      "ProductRequest": {
        "type": "object",
        "required": ["name", "price"],
        "properties": {
          "name": { "type": "string", "example": "Mechanical Keyboard" },
          "price": { "type": "number", "example": 750000 },
          "stock": { "type": "integer", "example": 12 },
          "category": { "type": "string", "example": "electronics" }
        }
      }
    }
  },
`