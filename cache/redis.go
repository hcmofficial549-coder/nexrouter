package cache

import (
"context"
"encoding/json"
"math"
"os"
"sync/atomic"
"time"

"github.com/redis/go-redis/v9"
)

const keyPrefix = "nrc:"

// RedisCache is a Redis-backed Cache implementation
type RedisCache struct {
client *redis.Client
hits   int64
misses int64
}

// NewRedisCache creates a Redis cache from a connection URL
func NewRedisCache(url string) *RedisCache {
opt, err := redis.ParseURL(url)
if err != nil {
return nil
}
return &RedisCache{client: redis.NewClient(opt)}
}

func (r *RedisCache) Name() string { return "redis" }

func (r *RedisCache) Get(key string) (interface{}, bool) {
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
data, err := r.client.Get(ctx, keyPrefix+key).Bytes()
if err != nil {
atomic.AddInt64(&r.misses, 1)
return nil, false
}
var out interface{}
if err := json.Unmarshal(data, &out); err != nil {
atomic.AddInt64(&r.misses, 1)
return nil, false
}
atomic.AddInt64(&r.hits, 1)
return out, true
}

func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) {
data, err := json.Marshal(value)
if err != nil {
return
}
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
r.client.Set(ctx, keyPrefix+key, data, ttl)
}

func (r *RedisCache) Delete(key string) {
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
r.client.Del(ctx, keyPrefix+key)
}

func (r *RedisCache) DeletePrefix(prefix string) {
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
iter := r.client.Scan(ctx, 0, keyPrefix+prefix+"*", 100).Iterator()
for iter.Next(ctx) {
r.client.Del(ctx, iter.Val())
}
}

func (r *RedisCache) Stats() map[string]interface{} {
h := atomic.LoadInt64(&r.hits)
ms := atomic.LoadInt64(&r.misses)
total := h + ms
rate := float64(0)
if total > 0 {
rate = float64(h) / float64(total) * 100
}
return map[string]interface{}{
"engine":           "redis",
"hits":             h,
"misses":           ms,
"hit_rate_percent": math.Round(rate*100) / 100,
}
}

// NewFromEnv returns Redis cache when available, otherwise memory cache.
// The app NEVER crashes because of cache - always falls back gracefully.
func NewFromEnv() Cache {
url := os.Getenv("REDIS_PRIVATE_URL")
if url == "" {
url = os.Getenv("REDIS_URL")
}
if url != "" {
if rc := NewRedisCache(url); rc != nil {
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
if err := rc.client.Ping(ctx).Err(); err == nil {
return rc
}
}
}
return NewMemoryCache(1000)
}