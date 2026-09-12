package chat

// FindClient mencari client online by exact name (first match).
func (h *Hub) FindClient(name string) *Client {
h.mu.RLock()
defer h.mu.RUnlock()
for _, clients := range h.rooms {
for c := range clients {
if c.Name == name {
return c
}
}
}
return nil
}

// SendPM mengirim private message.
// v1.5: SELALU dipersist ke SQLite (async messaging).
// Return true jika target online dan pesan terkirim live.
func (h *Hub) SendPM(from *Client, toName, text string) bool {
if toName == from.Name {
return false
}
m := Message{
Type:     "pm",
From:     from.Name,
To:       toName,
Text:     text,
Time:     nowStr(),
Verified: from.Verified,
Admin:    from.IsAdmin,
}
m.Room = DMKey(from.Name, toName)
go SaveMessage(m)

b := m.bytes()

// echo ke pengirim (selalu)
select {
case from.Send <- b:
default:
}

// delivery live jika target online
target := h.FindClient(toName)
if target == nil {
return false
}
select {
case target.Send <- b:
default:
}
return true
}

// NotifyClient mengirim system message ke satu client
func (h *Hub) NotifyClient(c *Client, text string) {
m := Message{Type: "system", From: "system", Text: text, Time: nowStr()}
select {
case c.Send <- m.bytes():
default:
}
}

// BroadcastSystem mengirim system message ke seluruh room
func (h *Hub) BroadcastSystem(room, text string) {
h.store(Message{Type: "system", From: "system", Room: room, Text: text, Time: nowStr()})
}

// Kick mengirim control frame __KICK__ ke target.
// Write pump target yang menutup koneksi (channel ownership aman).
func (h *Hub) Kick(name, reason string) bool {
target := h.FindClient(name)
if target == nil {
return false
}
select {
case target.Send <- []byte("__KICK__" + reason):
return true
default:
return false
}
}