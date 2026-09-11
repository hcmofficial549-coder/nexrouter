package events

import (
"sync"
"time"

"github.com/nexrouter/nexrouter/core"
)

// Event is a single realtime event
type Event struct {
Type      string `json:"type"`
Method    string `json:"method,omitempty"`
Path      string `json:"path,omitempty"`
Status    int    `json:"status,omitempty"`
LatencyMS int64  `json:"latency_ms,omitempty"`
IP        string `json:"ip,omitempty"`
Time      string `json:"time"`
}

// Hub is a pub/sub broadcaster: middleware publishes, SSE clients subscribe
type Hub struct {
mu   sync.RWMutex
subs map[chan Event]struct{}
buf  int
}

// NewHub creates a hub; buf = per-subscriber channel buffer
func NewHub(buf int) *Hub {
return &Hub{subs: make(map[chan Event]struct{}), buf: buf}
}

// Subscribe returns a receive-only event channel + unsubscribe function
func (h *Hub) Subscribe() (<-chan Event, func()) {
ch := make(chan Event, h.buf)
h.mu.Lock()
h.subs[ch] = struct{}{}
h.mu.Unlock()
unsub := func() {
h.mu.Lock()
if _, ok := h.subs[ch]; ok {
delete(h.subs, ch)
close(ch)
}
h.mu.Unlock()
}
return ch, unsub
}

// Publish broadcasts an event to all subscribers (non-blocking:
// slow consumers drop events instead of blocking the API)
func (h *Hub) Publish(ev Event) {
h.mu.RLock()
defer h.mu.RUnlock()
for ch := range h.subs {
select {
case ch <- ev:
default:
}
}
}

// Subscribers returns active subscriber count
func (h *Hub) Subscribers() int {
h.mu.RLock()
defer h.mu.RUnlock()
return len(h.subs)
}

// FeedMiddleware publishes every HTTP request as a realtime event
func FeedMiddleware(hub *Hub) core.HandlerFunc {
return func(c *core.Context) {
start := time.Now()
c.Next()
path := c.Request.URL.Path
if path == "/api/v1/events/stream" || path == "/monitor" || path == "/favicon.ico" {
return
}
hub.Publish(Event{
Type:      "request",
Method:    c.Request.Method,
Path:      path,
Status:    c.StatusCode(),
LatencyMS: time.Since(start).Milliseconds(),
IP:        c.ClientIP(),
Time:      time.Now().Format("15:04:05"),
})
}
}