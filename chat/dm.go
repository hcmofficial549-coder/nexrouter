package chat

import (
"sort"
"strings"
)

// sanitizeName makes a name safe for use inside a DM room key
func sanitizeName(name string) string {
n := strings.ToLower(strings.TrimSpace(name))
n = strings.ReplaceAll(n, ":", "-")
return n
}

// DMKey builds a canonical, order-independent DM room key:
// DMKey("Budi", "John") == DMKey("john", "BUDI") == "dm:budi:john"
func DMKey(a, b string) string {
parts := []string{sanitizeName(a), sanitizeName(b)}
sort.Strings(parts)
return "dm:" + parts[0] + ":" + parts[1]
}

// DMHistory loads persisted private messages between two users
func DMHistory(userA, userB string, limit int) []Message {
return LoadHistory(DMKey(userA, userB), limit)
}