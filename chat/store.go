package chat

import (
"database/sql"
"log"
"os"
"path/filepath"

_ "modernc.org/sqlite"
)

// DB is the chat persistence connection (nil = disabled, memory-only)
var DB *sql.DB

// InitStore opens SQLite for chat history persistence
func InitStore(dbPath string) error {
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
_, err = DB.Exec(`
CREATE TABLE IF NOT EXISTS messages (
id INTEGER PRIMARY KEY AUTOINCREMENT,
room TEXT NOT NULL,
sender TEXT NOT NULL,
mtype TEXT NOT NULL DEFAULT 'message',
text TEXT NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room, id);
`)
return err
}

// SaveMessage persists one chat message (called async via go)
func SaveMessage(m Message) {
if DB == nil {
return
}
if m.Type != "message" {
return
}
_, err := DB.Exec(
"INSERT INTO messages (room, sender, mtype, text) VALUES (?, ?, ?, ?)",
m.Room, m.From, m.Type, m.Text,
)
if err != nil {
log.Printf("[nexchat] save message failed: %v", err)
}
}

// LoadHistory loads persisted history for a room (chronological order)
func LoadHistory(room string, limit int) []Message {
if DB == nil {
return nil
}
rows, err := DB.Query(
"SELECT sender, text, strftime('%H:%M:%S', created_at) FROM messages WHERE room = ? ORDER BY id DESC LIMIT ?",
room, limit,
)
if err != nil {
return nil
}
defer rows.Close()
var rev []Message
for rows.Next() {
var m Message
if err := rows.Scan(&m.ID, &m.From, &m.Text, &m.Time); err != nil {
continue
}
m.Type = "message"
m.Room = room
rev = append(rev, m)
}
out := make([]Message, 0, len(rev))
for i := len(rev) - 1; i >= 0; i-- {
out = append(out, rev[i])
}
return out
}

// CloseStore closes the persistence connection
func CloseStore() {
if DB != nil {
DB.Close()
}
}