package cache

import (
"math"
"strings"
"sync"
"time"
)

// Cache is the caching layer interface
type Cache interface {
Get(key string) (interface{}, bool)
Set(key string, value interface{}, ttl time.Duration)
Delete(key string)
DeletePrefix(prefix string)
Stats() map[string]interface{}
Name() string
}

// MemoryCache is an in-process cache with TTL and LRU-ish eviction
type MemoryCache struct {
mu      sync.Mutex
items   map[string]memItem
maxSize int
hits    int64
misses  int64
}

type memItem struct {
value     interface{}
expiresAt time.Time
}

// NewMemoryCache creates a memory cache with background cleaner
func NewMemoryCache(maxSize int) *MemoryCache {
m := &MemoryCache{
items:   make(map[string]memItem),
maxSize: maxSize,
}
go m.cleanupLoop()
return m
}

func (m *MemoryCache) Name() string { return "memory" }

func (m *MemoryCache) Get(key string) (interface{}, bool) {
m.mu.Lock()
defer m.mu.Unlock()
item, ok := m.items[key]
if !ok || time.Now().After(item.expiresAt) {
if ok {
delete(m.items, key)
}
m.misses++
return nil, false
}
m.hits++
return item.value, true
}

func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
m.mu.Lock()
defer m.mu.Unlock()
if len(m.items) >= m.maxSize {
m.evictOne()
}
m.items[key] = memItem{value: value, expiresAt: time.Now().Add(ttl)}
}

func (m *MemoryCache) Delete(key string) {
m.mu.Lock()
defer m.mu.Unlock()
delete(m.items, key)
}

func (m *MemoryCache) DeletePrefix(prefix string) {
m.mu.Lock()
defer m.mu.Unlock()
for k := range m.items {
if strings.HasPrefix(k, prefix) {
delete(m.items, k)
}
}
}

func (m *MemoryCache) evictOne() {
var oldestKey string
var oldestTime time.Time
for k, v := range m.items {
if oldestKey == "" || v.expiresAt.Before(oldestTime) {
oldestKey = k
oldestTime = v.expiresAt
}
}
if oldestKey != "" {
delete(m.items, oldestKey)
}
}

func (m *MemoryCache) cleanupLoop() {
for {
time.Sleep(time.Minute)
m.mu.Lock()
now := time.Now()
for k, v := range m.items {
if now.After(v.expiresAt) {
delete(m.items, k)
}
}
m.mu.Unlock()
}
}

func (m *MemoryCache) Stats() map[string]interface{} {
m.mu.Lock()
defer m.mu.Unlock()
total := m.hits + m.misses
rate := float64(0)
if total > 0 {
rate = float64(m.hits) / float64(total) * 100
}
return map[string]interface{}{
"engine":           "memory",
"entries":          len(m.items),
"hits":             m.hits,
"misses":           m.misses,
"hit_rate_percent": math.Round(rate*100) / 100,
}
}