package docs

const specPart2 = `
  "paths": {
    "/health": {
      "get": {
        "tags": ["system"],
        "summary": "Health check",
        "description": "Returns service status, version, and active engines (database/auth/cache).",
        "responses": {
          "200": { "description": "Service healthy" }
        }
      }
    },
    "/api/v1/auth/register": {
      "post": {
        "tags": ["auth"],
        "summary": "Register new user",
        "description": "Creates an account (password hashed with bcrypt) and returns a JWT token - auto login.",
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "$ref": "#/components/schemas/RegisterRequest" } } }
        },
        "responses": {
          "201": { "description": "Registered, token returned", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/TokenResponse" } } } },
          "400": { "description": "Validation error", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Error" } } } },
          "409": { "description": "Email already registered" }
        }
      }
    },
    "/api/v1/auth/login": {
      "post": {
        "tags": ["auth"],
        "summary": "Login and get JWT token",
        "description": "Demo account: john@example.com / password123. Token valid for 24 hours.",
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "$ref": "#/components/schemas/LoginRequest" } } }
        },
        "responses": {
          "200": { "description": "Login success", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/TokenResponse" } } } },
          "401": { "description": "Invalid email or password" }
        }
      }
    },
    "/api/v1/auth/me": {
      "get": {
        "tags": ["auth"],
        "summary": "Current user profile (from token)",
        "security": [{ "bearerAuth": [] }],
        "responses": {
          "200": { "description": "Profile from JWT claims" },
          "401": { "description": "Missing or invalid token" }
        }
      }
    },
    "/api/v1/users": {
      "get": {
        "tags": ["users"],
        "summary": "List all users",
        "responses": {
          "200": {
            "description": "User list",
            "content": { "application/json": { "schema": { "type": "object", "properties": { "success": { "type": "boolean" }, "count": { "type": "integer" }, "data": { "type": "array", "items": { "$ref": "#/components/schemas/User" } } } } } }
          }
        }
      }
    },
    "/api/v1/users/{id}": {
      "get": {
        "tags": ["users"],
        "summary": "Get user by ID",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "integer" }, "example": 1 }],
        "responses": {
          "200": { "description": "User found" },
          "404": { "description": "User not found" }
        }
      }
    },
    "/api/v1/products": {
      "get": {
        "tags": ["products"],
        "summary": "List products (cached 30s)",
        "description": "Check the X-Cache response header: HIT = served from cache, MISS = from SQLite. Optional category filter.",
        "parameters": [{ "name": "category", "in": "query", "required": false, "schema": { "type": "string" }, "example": "electronics" }],
        "responses": {
          "200": {
            "description": "Product list",
            "content": { "application/json": { "schema": { "type": "object", "properties": { "success": { "type": "boolean" }, "count": { "type": "integer" }, "data": { "type": "array", "items": { "$ref": "#/components/schemas/Product" } } } } } }
          }
        }
      },
      "post": {
        "tags": ["products"],
        "summary": "Create product",
        "security": [{ "bearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ProductRequest" } } }
        },
        "responses": {
          "201": { "description": "Created (cache invalidated)" },
          "400": { "description": "Validation error" },
          "401": { "description": "Token required" }
        }
      }
    },
    "/api/v1/products/{id}": {
      "get": {
        "tags": ["products"],
        "summary": "Get product by ID",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }],
        "responses": {
          "200": { "description": "Product found" },
          "404": { "description": "Not found" }
        }
      },
      "put": {
        "tags": ["products"],
        "summary": "Update product (partial update supported)",
        "security": [{ "bearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ProductRequest" } } }
        },
        "responses": {
          "200": { "description": "Updated (cache invalidated)" },
          "401": { "description": "Token required" },
          "404": { "description": "Not found" }
        }
      },
      "delete": {
        "tags": ["products"],
        "summary": "Delete product",
        "security": [{ "bearerAuth": [] }],
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "integer" } }],
        "responses": {
          "200": { "description": "Deleted (cache invalidated)" },
          "401": { "description": "Token required" },
          "404": { "description": "Not found" }
        }
      }
    },
    "/api/v1/stats": {
      "get": {
        "tags": ["system"],
        "summary": "Inventory statistics (cached 30s)",
        "security": [{ "bearerAuth": [] }],
        "responses": {
          "200": { "description": "Aggregate stats", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Stats" } } } },
          "401": { "description": "Token required" }
        }
      }
    },
    "/api/v1/cache/stats": {
      "get": {
        "tags": ["system"],
        "summary": "Cache engine statistics",
        "description": "Shows cache engine (memory/redis), entries, hits, misses, and hit rate.",
        "responses": {
          "200": { "description": "Cache stats" }
        }
      }
    },
    "/api/v1/events/stream": {
      "get": {
        "tags": ["system"],
        "summary": "Realtime event stream (SSE)",
        "description": "Server-Sent Events stream of every API request. Open with EventSource in JS or: curl -N. Also try the visual version at /monitor.",
        "responses": {
          "200": { "description": "text/event-stream", "content": { "text/event-stream": { "schema": { "type": "string" } } } }
        }
      }
    }
  }
}
`