package database

import "database/sql"

// UserAuth includes password hash - INTERNAL ONLY
type UserAuth struct {
ID           int
Name         string
Email        string
PasswordHash string
}

// MigrateAuth adds password_hash column if missing (idempotent)
func MigrateAuth() error {
var colCount int
err := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'password_hash'").Scan(&colCount)
if err != nil {
return err
}
if colCount == 0 {
if _, err := DB.Exec("ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT ''"); err != nil {
return err
}
}
return nil
}

// CreateUserWithPassword creates a user with bcrypt hash
func CreateUserWithPassword(name, email, passwordHash string) (*User, error) {
res, err := DB.Exec("INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)", name, email, passwordHash)
if err != nil {
return nil, err
}
id, _ := res.LastInsertId()
return &User{ID: int(id), Name: name, Email: email}, nil
}

// GetUserAuthByEmail finds a user with password hash for login
func GetUserAuthByEmail(email string) (*UserAuth, error) {
var u UserAuth
err := DB.QueryRow("SELECT id, name, email, password_hash FROM users WHERE email = ?", email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash)
if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}
return &u, nil
}

// SetDefaultPasswords sets password for legacy users without one
func SetDefaultPasswords(hash string) int64 {
res, err := DB.Exec("UPDATE users SET password_hash = ? WHERE password_hash = '' OR password_hash IS NULL", hash)
if err != nil {
return 0
}
n, _ := res.RowsAffected()
return n
}