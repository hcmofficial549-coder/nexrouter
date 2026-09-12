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

// DMThread is one DM conversation entry for the sidebar inbox
type DMThread struct {
With string `json:"with"`
Last string `json:"last"`
}

// DMThreads lists the DM counterparts of a user from persisted PMs,
// most recent conversation first. Original display names preserved
// (diambil dari kolom sender/recipient, bukan parse room key).
func DMThreads(user string) []DMThread {
if DB == nil {
return nil
}
me := strings.ToLower(strings.TrimSpace(user))
rows, err := DB.Query(
"SELECT CASE WHEN LOWER(sender) = ? THEN recipient ELSE sender END AS other, strftime('%H:%M:%S', MAX(created_at)) AS last FROM messages WHERE mtype = 'pm' AND (LOWER(sender) = ? OR LOWER(recipient) = ?) GROUP BY other ORDER BY MAX(created_at) DESC LIMIT 20",
me, me, me,
)
if err != nil {
return nil
}
defer rows.Close()
out := []DMThread{}
for rows.Next() {
var t DMThread
if err := rows.Scan(&t.With, &t.Last); err != nil {
continue
}
out = append(out, t)
}
return out
}