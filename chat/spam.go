package chat

import (
"sync"
"time"
)

// SpamGuard limits message rate per client with escalating mutes.
type SpamGuard struct {
mu       sync.Mutex
window   time.Duration
maxMsgs  int
buckets  map[*Client]*spamBucket
strikes  map[*Client]int
mutedTil map[*Client]time.Time
}

type spamBucket struct {
count       int
windowStart time.Time
}

// NewSpamGuard allows maxMsgs per window. Violators get escalating
// mutes: strike x 30 seconds (30s, 60s, 90s, ...).
func NewSpamGuard(maxMsgs int, window time.Duration) *SpamGuard {
return &SpamGuard{
window:   window,
maxMsgs:  maxMsgs,
buckets:  make(map[*Client]*spamBucket),
strikes:  make(map[*Client]int),
mutedTil: make(map[*Client]time.Time),
}
}

// Allow reports whether the client may send a message now.
// When false, the second return value is remaining mute seconds.
func (s *SpamGuard) Allow(c *Client) (bool, int) {
s.mu.Lock()
defer s.mu.Unlock()

now := time.Now()

// sedang di-mute?
if til, ok := s.mutedTil[c]; ok {
if now.Before(til) {
return false, int(til.Sub(now).Seconds()) + 1
}
delete(s.mutedTil, c)
}

// sliding window sederhana
b := s.buckets[c]
if b == nil || now.Sub(b.windowStart) >= s.window {
s.buckets[c] = &spamBucket{count: 1, windowStart: now}
return true, 0
}
b.count++
if b.count <= s.maxMsgs {
return true, 0
}

// pelanggaran: strike + mute eskalatif
s.strikes[c]++
dur := time.Duration(s.strikes[c]) * 30 * time.Second
s.mutedTil[c] = now.Add(dur)
s.buckets[c] = &spamBucket{count: 0, windowStart: now}
return false, int(dur.Seconds())
}

// Cleanup forgets a disconnected client (anti memory leak)
func (s *SpamGuard) Cleanup(c *Client) {
s.mu.Lock()
defer s.mu.Unlock()
delete(s.buckets, c)
delete(s.strikes, c)
delete(s.mutedTil, c)
}

// Unmute forcibly removes a mute on a client and resets their strikes.
func (s *SpamGuard) Unmute(c *Client) {
s.mu.Lock()
defer s.mu.Unlock()
delete(s.mutedTil, c)
delete(s.strikes, c)
delete(s.buckets, c)
}