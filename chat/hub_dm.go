package chat

// FindClient mencari client online by exact name (first match).
// Catatan: nama duplikat mungkin; PM pergi ke yang pertama ketemu.
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

// SendPM mengirim private message ke target (dan echo ke pengirim).
// PM bersifat ephemeral: tidak disimpan di history, tidak persisten.
func (h *Hub) SendPM(from *Client, toName, text string) bool {
target := h.FindClient(toName)
if target == nil || target == from {
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
b := m.bytes()
for _, t := range []*Client{target, from} {
select {
case t.Send <- b:
default:
}
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
// Write pump target yang menutup koneksi; read pump-nya sendiri
// yang cleanup (channel ownership: hanya pemilik yang close).
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