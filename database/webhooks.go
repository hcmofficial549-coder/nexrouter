package database

import (
"database/sql"
"strings"
)

// Webhook is an event subscription
type Webhook struct {
ID     int    `json:"id"`
URL    string `json:"url"`
Events string `json:"events"`
Secret string `json:"secret"`
Active bool   `json:"active"`
}

// MigrateWebhooks creates the webhooks table (idempotent)
func MigrateWebhooks() error {
schema := `
CREATE TABLE IF NOT EXISTS webhooks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
url TEXT NOT NULL,
events TEXT NOT NULL DEFAULT '*',
secret TEXT NOT NULL,
active INTEGER NOT NULL DEFAULT 1,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
_, err := DB.Exec(schema)
return err
}

// CreateWebhook inserts a subscription
func CreateWebhook(url, events, secret string) (*Webhook, error) {
res, err := DB.Exec(
"INSERT INTO webhooks (url, events, secret) VALUES (?, ?, ?)",
url, events, secret,
)
if err != nil {
return nil, err
}
id, _ := res.LastInsertId()
return &Webhook{ID: int(id), URL: url, Events: events, Secret: secret, Active: true}, nil
}

// ListWebhooks returns all subscriptions
func ListWebhooks() ([]Webhook, error) {
rows, err := DB.Query("SELECT id, url, events, secret, active FROM webhooks ORDER BY id")
if err != nil {
return nil, err
}
defer rows.Close()
items := []Webhook{}
for rows.Next() {
var w Webhook
var active int
if err := rows.Scan(&w.ID, &w.URL, &w.Events, &w.Secret, &active); err != nil {
return nil, err
}
w.Active = active == 1
items = append(items, w)
}
return items, rows.Err()
}

// GetWebhook returns one subscription
func GetWebhook(id int) (*Webhook, error) {
var w Webhook
var active int
err := DB.QueryRow(
"SELECT id, url, events, secret, active FROM webhooks WHERE id = ?", id,
).Scan(&w.ID, &w.URL, &w.Events, &w.Secret, &active)
if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}
w.Active = active == 1
return &w, nil
}

// DeleteWebhook removes a subscription
func DeleteWebhook(id int) (bool, error) {
res, err := DB.Exec("DELETE FROM webhooks WHERE id = ?", id)
if err != nil {
return false, err
}
n, _ := res.RowsAffected()
return n > 0, nil
}

// ListActiveWebhooksForEvent returns active hooks subscribed to event
func ListActiveWebhooksForEvent(event string) ([]Webhook, error) {
rows, err := DB.Query("SELECT id, url, events, secret FROM webhooks WHERE active = 1")
if err != nil {
return nil, err
}
defer rows.Close()
items := []Webhook{}
for rows.Next() {
var w Webhook
if err := rows.Scan(&w.ID, &w.URL, &w.Events, &w.Secret); err != nil {
return nil, err
}
for _, e := range strings.Split(w.Events, ",") {
e = strings.TrimSpace(e)
if e == "*" || e == event {
items = append(items, w)
break
}
}
}
return items, rows.Err()
}