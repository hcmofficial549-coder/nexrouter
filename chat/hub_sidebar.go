package chat

import "sort"

// RoomsFull lists all known rooms: rooms with online clients PLUS
// rooms that still have history, sorted by most recent activity.
func (h *Hub) RoomsFull() []RoomInfo {
h.mu.RLock()
defer h.mu.RUnlock()
seen := map[string]bool{}
out := []RoomInfo{}
for name, clients := range h.rooms {
out = append(out, RoomInfo{Name: name, Online: len(clients), LastAct: h.lastAct[name]})
seen[name] = true
}
for name, act := range h.lastAct {
if !seen[name] {
out = append(out, RoomInfo{Name: name, Online: 0, LastAct: act})
seen[name] = true
}
}
if !seen["general"] {
out = append(out, RoomInfo{Name: "general", Online: 0})
}
sort.Slice(out, func(i, j int) bool { return out[i].LastAct > out[j].LastAct })
return out
}