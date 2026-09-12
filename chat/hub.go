package chat

import (
"encoding/json"
"sync"
"time"
)

// Message is one chat event
type Message struct {
Type string `json:"type"`
From string `json:"from"`
Room string `json:"room"`
Text string `json:"text"`
Time string `json:"time"`
	Verified bool `json:"verified,omitempty"`
	To    string `json:"to,omitempty"`
	Admin bool   `json:"admin,omitempty"`
}

// RoomInfo describes a room for the REST API
type RoomInfo struct {
Name   string `json:"name"`
Online int    `json:"online"`
}

// Client is one connected websocket user
type Client struct {
Name string
Room string
Send chan []byte
	Verified bool
	IsAdmin bool
}

// Hub manages rooms, clients, broadcast and history
type Hub struct {
mu      sync.RWMutex
rooms   map[string]map[*Client]bool
history map[string][]Message
}

const historyLimit = 50

// NewHub creates a chat hub
func NewHub() *Hub {
return &Hub{
rooms:   make(map[string]map[*Client]bool),
history: make(map[string][]Message),
}
}

func nowStr() string { return time.Now().Format("15:04:05") }

func (m Message) bytes() []byte {
b, _ := json.Marshal(m)
return b
}

// Join registers a client and announces the join
func (h *Hub) Join(name, room string, send chan []byte) *Client {
c := &Client{Name: name, Room: room, Send: send}
h.mu.Lock()
if h.rooms[room] == nil {
h.rooms[room] = make(map[*Client]bool)
}
h.rooms[room][c] = true
h.mu.Unlock()
h.store(Message{Type: "join", From: name, Room: room, Text: name + " joined the room", Time: nowStr()})
return c
}

// Leave removes a client and announces the leave
func (h *Hub) Leave(c *Client) {
h.mu.Lock()
if clients, ok := h.rooms[c.Room]; ok {
if _, exists := clients[c]; exists {
delete(clients, c)
if len(clients) == 0 {
delete(h.rooms, c.Room)
}
}
}
h.mu.Unlock()
h.store(Message{Type: "leave", From: c.Name, Room: c.Room, Text: c.Name + " left the room", Time: nowStr()})
}

// Broadcast sends a chat message from a client to the whole room
func (h *Hub) Broadcast(c *Client, text string) {
h.store(Message{Type: "message", From: c.Name, Room: c.Room, Text: text, Time: nowStr(), Verified: c.Verified, Admin: c.IsAdmin})
}

// store appends to history and fans out to room members (non-blocking)
func (h *Hub) store(m Message) {
b := m.bytes()
	go SaveMessage(m)
h.mu.Lock()
hist := append(h.history[m.Room], m)
if len(hist) > historyLimit {
hist = hist[len(hist)-historyLimit:]
}
h.history[m.Room] = hist
clients := h.rooms[m.Room]
targets := make([]*Client, 0, len(clients))
for c := range clients {
targets = append(targets, c)
}
h.mu.Unlock()

for _, c := range targets {
select {
case c.Send <- b:
default:
}
}
}

// History returns recent messages of a room
func (h *Hub) History(room string) []Message {
h.mu.RLock()
defer h.mu.RUnlock()
out := make([]Message, len(h.history[room]))
copy(out, h.history[room])
return out
}

// Rooms lists all rooms with online counts
func (h *Hub) Rooms() []RoomInfo {
h.mu.RLock()
defer h.mu.RUnlock()
out := []RoomInfo{}
found := false
for name, clients := range h.rooms {
out = append(out, RoomInfo{Name: name, Online: len(clients)})
if name == "general" {
found = true
}
}
if !found {
out = append(out, RoomInfo{Name: "general", Online: 0})
}
return out
}

// Online returns online count in a room
func (h *Hub) Online(room string) int {
h.mu.RLock()
defer h.mu.RUnlock()
return len(h.rooms[room])
}