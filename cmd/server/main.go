package main

import (
"log"
"net/http"
"os"
"os/signal"
"strconv"
"time"

"github.com/nexrouter/nexrouter/auth"
"github.com/nexrouter/nexrouter/core"
"github.com/nexrouter/nexrouter/database"
"github.com/nexrouter/nexrouter/middleware"
)

var jwtSecret = "nexrouter-dev-secret-change-me"

func main() {
if s := os.Getenv("JWT_SECRET"); s != "" {
jwtSecret = s
}

dbPath := os.Getenv("DB_PATH")
if dbPath == "" {
dbPath = "data/nexrouter.db"
}
if err := database.Init(dbPath); err != nil {
log.Fatalf("[nexrouter] database init failed: %v", err)
}
if err := database.MigrateAuth(); err != nil {
log.Fatalf("[nexrouter] auth migration failed: %v", err)
}
log.Printf("[nexrouter] SQLite ready at: %s", dbPath)

if hash, err := auth.HashPassword("password123"); err == nil {
if n := database.SetDefaultPasswords(hash); n > 0 {
log.Printf("[nexrouter] set default password for %d legacy user(s)", n)
}
}

r := core.New()
r.Use(middleware.Recovery())
r.Use(middleware.Logger())
r.Use(middleware.CORS())

r.GET("/health", func(c *core.Context) {
c.JSON(http.StatusOK, core.H{"status": "ok", "version": "1.3.0", "database": "sqlite", "auth": "jwt"})
})

r.GET("/", func(c *core.Context) {
c.JSON(http.StatusOK, core.H{
"message": "Welcome to nexrouter!",
"version": "1.3.0",
"demo_login": core.H{"email": "john@example.com", "password": "password123"},
})
})

api := r.Group("/api/v1")
api.POST("/auth/register", registerHandler)
api.POST("/auth/login", loginHandler)
api.GET("/users", listUsers)
api.GET("/users/:id", getUser)
api.GET("/products", listProducts)
api.GET("/products/:id", getProduct)

prot := r.Group("/api/v1")
prot.Use(auth.JWT(jwtSecret))
prot.GET("/auth/me", meHandler)
prot.GET("/stats", getStats)
prot.POST("/products", createProduct)
prot.PUT("/products/:id", updateProduct)
prot.DELETE("/products/:id", deleteProduct)

go func() {
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
<-quit
log.Println("[nexrouter] Shutting down...")
database.Close()
os.Exit(0)
}()

port := os.Getenv("PORT")
if port == "" {
port = "8080"
}
log.Printf("[nexrouter] v1.3.0 (SQLite + JWT) starting on :%s", port)
if err := r.Run(":" + port); err != nil {
log.Fatal(err)
}
}

func registerHandler(c *core.Context) {
var input struct {
Name     string `json:"name"`
Email    string `json:"email"`
Password string `json:"password"`
}
if err := c.BindJSON(&input); err != nil {
c.BadRequest("invalid request body")
return
}
if input.Name == "" || input.Email == "" {
c.BadRequest("name and email are required")
return
}
if len(input.Password) < 6 {
c.BadRequest("password must be at least 6 characters")
return
}
hash, err := auth.HashPassword(input.Password)
if err != nil {
c.InternalError("failed to hash password")
return
}
u, err := database.CreateUserWithPassword(input.Name, input.Email, hash)
if err != nil {
c.JSON(http.StatusConflict, core.H{"error": "email already registered"})
return
}
token, err := auth.GenerateToken(strconv.Itoa(u.ID), u.Email, u.Name, jwtSecret, 24*time.Hour)
if err != nil {
c.InternalError("failed to generate token")
return
}
c.Created(core.H{"success": true, "user": u, "token": token, "expires_in": "24h"})
}

func loginHandler(c *core.Context) {
var input struct {
Email    string `json:"email"`
Password string `json:"password"`
}
if err := c.BindJSON(&input); err != nil {
c.BadRequest("invalid request body")
return
}
u, err := database.GetUserAuthByEmail(input.Email)
if err != nil {
c.InternalError(err.Error())
return
}
if u == nil || !auth.CheckPassword(u.PasswordHash, input.Password) {
c.Unauthorized("invalid email or password")
return
}
token, err := auth.GenerateToken(strconv.Itoa(u.ID), u.Email, u.Name, jwtSecret, 24*time.Hour)
if err != nil {
c.InternalError("failed to generate token")
return
}
c.JSON(http.StatusOK, core.H{
"success":    true,
"token":      token,
"expires_in": "24h",
"user":       core.H{"id": u.ID, "name": u.Name, "email": u.Email},
})
}

func meHandler(c *core.Context) {
claims := auth.GetClaims(c)
if claims == nil {
c.Unauthorized("unauthorized")
return
}
c.JSON(http.StatusOK, core.H{
"success":          true,
"user":             core.H{"id": claims.Sub, "name": claims.Name, "email": claims.Email},
"token_expires_at": time.Unix(claims.Exp, 0).Format(time.RFC3339),
})
}
func listUsers(c *core.Context) {
users, err := database.ListUsers()
if err != nil {
c.InternalError(err.Error())
return
}
c.JSON(http.StatusOK, core.H{"success": true, "data": users, "count": len(users)})
}

func getUser(c *core.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.BadRequest("invalid user id")
return
}
u, err := database.GetUser(id)
if err != nil {
c.InternalError(err.Error())
return
}
if u == nil {
c.NotFound("user not found")
return
}
c.JSON(http.StatusOK, u)
}

func listProducts(c *core.Context) {
products, err := database.ListProducts(c.Query("category"))
if err != nil {
c.InternalError(err.Error())
return
}
c.JSON(http.StatusOK, core.H{"success": true, "data": products, "count": len(products)})
}

func getProduct(c *core.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.BadRequest("invalid product id")
return
}
p, err := database.GetProduct(id)
if err != nil {
c.InternalError(err.Error())
return
}
if p == nil {
c.NotFound("product not found")
return
}
c.JSON(http.StatusOK, p)
}

func createProduct(c *core.Context) {
var input database.Product
if err := c.BindJSON(&input); err != nil {
c.BadRequest("invalid request body")
return
}
if input.Name == "" {
c.BadRequest("name is required")
return
}
if input.Price <= 0 {
c.BadRequest("price must be greater than 0")
return
}
p, err := database.CreateProduct(&input)
if err != nil {
c.InternalError(err.Error())
return
}
createdBy := ""
if cl := auth.GetClaims(c); cl != nil {
createdBy = cl.Email
}
c.Created(core.H{"success": true, "data": p, "created_by": createdBy})
}

func updateProduct(c *core.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.BadRequest("invalid product id")
return
}
var input struct {
Name     *string  `json:"name"`
Price    *float64 `json:"price"`
Stock    *int     `json:"stock"`
Category *string  `json:"category"`
}
if err := c.BindJSON(&input); err != nil {
c.BadRequest("invalid request body")
return
}
if input.Price != nil && *input.Price <= 0 {
c.BadRequest("price must be greater than 0")
return
}
p, err := database.UpdateProduct(id, &database.ProductUpdate{
Name:     input.Name,
Price:    input.Price,
Stock:    input.Stock,
Category: input.Category,
})
if err != nil {
c.InternalError(err.Error())
return
}
if p == nil {
c.NotFound("product not found")
return
}
c.JSON(http.StatusOK, core.H{"success": true, "data": p})
}

func deleteProduct(c *core.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.BadRequest("invalid product id")
return
}
deleted, err := database.DeleteProduct(id)
if err != nil {
c.InternalError(err.Error())
return
}
if !deleted {
c.NotFound("product not found")
return
}
c.JSON(http.StatusOK, core.H{"success": true, "message": "product deleted"})
}

func getStats(c *core.Context) {
s, err := database.GetStats()
if err != nil {
c.InternalError(err.Error())
return
}
c.JSON(http.StatusOK, s)
}