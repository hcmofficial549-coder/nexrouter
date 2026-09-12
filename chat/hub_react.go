package chat

import (
"encoding/json"
"sync"
"sync/atomic"
)

var (
msgCounter int64
reactionMu sync.Mutex
reactions  = map[int64]map[string]map[string]bool{}
)

// nextMsgID returns a unique message id for this session
func nextMsgID() int64 {
return atomic.AddInt64(&msgCounter, 1)
}

// SeedMsgCounter syncs the in-memory counter with max DB id,
// so new message ids never collide with persisted history ids.
func SeedMsgCounter() {
if DB == nil {
return
}
var maxID int64
if err := DB.QueryRow("SELECT COALESCE(MAX(id), 0) FROM messages").Scan(&maxID); err == nil {
atomic.StoreInt64(&msgCounter, maxID)
}
}

// AddReaction toggles a user's emoji on msgID.
// Returns snapshot: emoji -> list of users who reacted.
func AddReaction(msgID int64, emoji, user string) map[string][]string {
reactionMu.Lock()
defer reactionMu.Unlock()
if reactions[msgID] == nil {
reactions[msgID] = map[string]map[string]bool{}
}
if reactions[msgID][emoji] == nil {
reactions[msgID][emoji] = map[string]bool{}
}
if reactions[msgID][emoji][user] {
delete(reactions[msgID][emoji], user)
} else {
reactions[msgID][emoji][user] = true
}
snap := map[string][]string{}
for e, users := range reactions[msgID] {
list := []string{}
for u := range users {
list = append(list, u)
}
if len(list) > 0 {
snap[e] = list
}
}
return snap
}

// BroadcastReaction fans out a reaction update to a room
func (h *Hub) BroadcastReaction(room string, msgID int64, snapshot map[string][]string) {
payload := map[string]interface{}{
"type":      "react",
"msg_id":    msgID,
"reactions": snapshot,
}
b, err := json.Marshal(payload)
if err != nil {
return
}
h.mu.RLock()
clients := h.rooms[room]
targets := make([]*Client, 0, len(clients))
for c := range clients {
targets = append(targets, c)
}
h.mu.RUnlock()
for _, t := range targets {
select {
case t.Send <- b:
default:
}
}
}