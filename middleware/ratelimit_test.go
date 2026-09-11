package middleware

import "testing"

func TestRateLimiterBlocks(t *testing.T) {
rl := NewRateLimiter(1, 2)
if !rl.Allow("ip1") {
t.Error("1st request should pass")
}
if !rl.Allow("ip1") {
t.Error("2nd request should pass (burst 2)")
}
if rl.Allow("ip1") {
t.Error("3rd request should be blocked")
}
if !rl.Allow("ip2") {
t.Error("different IP should pass")
}
}