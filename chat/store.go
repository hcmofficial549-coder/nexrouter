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
if err != nil {
return err
}
// v1.5 migration: kolom recipient untuk DM (idempotent)
var colCount int
qerr := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('messages') WHERE name = 'recipient'").Scan(&colCount)
if qerr == nil && colCount == 0 {
if _, aerr := DB.Exec("ALTER TABLE messages ADD COLUMN recipient TEXT NOT NULL DEFAULT ''"); aerr != nil {
return aerr
}
log.Printf("[nexchat] migration: kolom recipient ditambahkan")
}
return nil
}

// SaveMessage persists chat messages AND private messages
func SaveMessage(m Message) {
if DB == nil {
return
}
if m.Type != "message" && m.Type != "pm" {
return
}
_, err := DB.Exec(
"INSERT INTO messages (room, sender, mtype, text, recipient) VALUES (?, ?, ?, ?, ?)",
m.Room, m.From, m.Type, m.Text, m.To,
)
if err != nil {
log.Printf("[nexchat] save message failed: %v", err)
}
}

// LoadHistory loads persisted history for a room OR dm key (chronological)
func LoadHistory(room string, limit int) []Message {
if DB == nil {
return nil
}
rows, err := DB.Query(
"SELECT id, sender, text, strftime('%H:%M:%S', created_at), mtype, recipient FROM messages WHERE room = ? ORDER BY id DESC LIMIT ?",
room, limit,
)
if err != nil {
return nil
}
defer rows.Close()
var rev []Message
for rows.Next() {
var m Message
if err := rows.Scan(&m.ID, &m.From, &m.Text, &m.Time, &m.Type, &m.To); err != nil {
continue
}
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