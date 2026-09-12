package chat

// BroadcastTyping sends an ephemeral typing event to the room
// (everyone except the sender). NOT stored in history, NOT persisted.
func (h *Hub) BroadcastTyping(c *Client) {
m := Message{Type: "typing", From: c.Name, Room: c.Room, Time: nowStr()}
b := m.bytes()

h.mu.RLock()
clients := h.rooms[c.Room]
targets := make([]*Client, 0, len(clients))
for t := range clients {
if t != c {
targets = append(targets, t)
}
}
h.mu.RUnlock()

for _, t := range targets {
select {
case t.Send <- b:
default:
}
}
}

// OnlineUsers returns the names of clients currently in a room
func (h *Hub) OnlineUsers(room string) []string {
h.mu.RLock()
defer h.mu.RUnlock()
out := []string{}
for c := range h.rooms[room] {
out = append(out, c.Name)
}
return out
}