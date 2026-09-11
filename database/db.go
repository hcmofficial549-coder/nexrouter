package database

import (
"database/sql"
"os"
"path/filepath"

_ "modernc.org/sqlite"
)

// DB is the global SQLite connection
var DB *sql.DB

// User model (JSON-safe, no password)
type User struct {
ID    int    `json:"id"`
Name  string `json:"name"`
Email string `json:"email"`
}

// Product model
type Product struct {
ID       int     `json:"id"`
Name     string  `json:"name"`
Price    float64 `json:"price"`
Stock    int     `json:"stock"`
Category string  `json:"category"`
}

// Stats summary
type Stats struct {
Users          int     `json:"users"`
Products       int     `json:"products"`
TotalStock     int     `json:"total_stock"`
InventoryValue float64 `json:"inventory_value"`
}

// Init opens SQLite and runs migrations + seed
func Init(dbPath string) error {
dir := filepath.Dir(dbPath)
if dir != "" && dir != "." {
if err := os.MkdirAll(dir, 0o755); err != nil {
return err
}
}
var err error
DB, err = sql.Open("sqlite", dbPath)
if err != nil {
return err
}
DB.SetMaxOpenConns(1)
if err := DB.Ping(); err != nil {
return err
}
if err := migrate(); err != nil {
return err
}
return seed()
}

func migrate() error {
schema := `
CREATE TABLE IF NOT EXISTS users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
email TEXT NOT NULL UNIQUE,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS products (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
price REAL NOT NULL,
stock INTEGER NOT NULL DEFAULT 0,
category TEXT NOT NULL DEFAULT 'general',
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
`
_, err := DB.Exec(schema)
return err
}

func seed() error {
var count int
if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
return err
}
if count == 0 {
if _, err := DB.Exec("INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com'), ('Jane Smith', 'jane@example.com')"); err != nil {
return err
}
}
if err := DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&count); err != nil {
return err
}
if count == 0 {
if _, err := DB.Exec("INSERT INTO products (name, price, stock, category) VALUES ('Laptop', 999.99, 10, 'electronics'), ('Mouse', 29.99, 50, 'electronics'), ('Desk Chair', 199.5, 8, 'furniture')"); err != nil {
return err
}
}
return nil
}

// ListUsers returns all users
func ListUsers() ([]User, error) {
rows, err := DB.Query("SELECT id, name, email FROM users ORDER BY id")
if err != nil {
return nil, err
}
defer rows.Close()
users := []User{}
for rows.Next() {
var u User
if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
return nil, err
}
users = append(users, u)
}
return users, rows.Err()
}

// GetUser returns one user by id
func GetUser(id int) (*User, error) {
var u User
err := DB.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name, &u.Email)
if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}
return &u, nil
}

// ListProducts returns products, optionally filtered by category
func ListProducts(category string) ([]Product, error) {
var rows *sql.Rows
var err error
if category != "" {
rows, err = DB.Query("SELECT id, name, price, stock, category FROM products WHERE category = ? ORDER BY id", category)
} else {
rows, err = DB.Query("SELECT id, name, price, stock, category FROM products ORDER BY id")
}
if err != nil {
return nil, err
}
defer rows.Close()
products := []Product{}
for rows.Next() {
var p Product
if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category); err != nil {
return nil, err
}
products = append(products, p)
}
return products, rows.Err()
}

// GetProduct returns one product by id
func GetProduct(id int) (*Product, error) {
var p Product
err := DB.QueryRow("SELECT id, name, price, stock, category FROM products WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category)
if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}
return &p, nil
}

// CreateProduct inserts a new product
func CreateProduct(p *Product) (*Product, error) {
if p.Category == "" {
p.Category = "general"
}
res, err := DB.Exec("INSERT INTO products (name, price, stock, category) VALUES (?, ?, ?, ?)", p.Name, p.Price, p.Stock, p.Category)
if err != nil {
return nil, err
}
id, _ := res.LastInsertId()
p.ID = int(id)
return p, nil
}

// ProductUpdate holds optional fields for partial update
type ProductUpdate struct {
Name     *string
Price    *float64
Stock    *int
Category *string
}

// UpdateProduct updates only provided fields
func UpdateProduct(id int, u *ProductUpdate) (*Product, error) {
p, err := GetProduct(id)
if err != nil || p == nil {
return p, err
}
if u.Name != nil {
p.Name = *u.Name
}
if u.Price != nil {
p.Price = *u.Price
}
if u.Stock != nil {
p.Stock = *u.Stock
}
if u.Category != nil {
p.Category = *u.Category
}
_, err = DB.Exec("UPDATE products SET name = ?, price = ?, stock = ?, category = ? WHERE id = ?", p.Name, p.Price, p.Stock, p.Category, id)
if err != nil {
return nil, err
}
return p, nil
}

// DeleteProduct removes a product
func DeleteProduct(id int) (bool, error) {
res, err := DB.Exec("DELETE FROM products WHERE id = ?", id)
if err != nil {
return false, err
}
n, _ := res.RowsAffected()
return n > 0, nil
}

// GetStats returns aggregate stats
func GetStats() (*Stats, error) {
s := &Stats{}
if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&s.Users); err != nil {
return nil, err
}
if err := DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&s.Products); err != nil {
return nil, err
}
err := DB.QueryRow("SELECT COALESCE(SUM(stock), 0), COALESCE(SUM(price * stock), 0) FROM products").Scan(&s.TotalStock, &s.InventoryValue)
if err != nil {
return nil, err
}
return s, nil
}

// Close closes the database
func Close() error {
if DB != nil {
return DB.Close()
}
return nil
}