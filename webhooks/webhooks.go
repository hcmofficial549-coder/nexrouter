package webhooks

import (
"bytes"
"crypto/hmac"
"crypto/sha256"
"encoding/hex"
"encoding/json"
"fmt"
"log"
"net/http"
"time"

"github.com/nexrouter/nexrouter/database"
)

// Payload is the webhook request body
type Payload struct {
Event string      `json:"event"`
Time  string      `json:"time"`
Data  interface{} `json:"data"`
}

var client = &http.Client{Timeout: 10 * time.Second}

// Dispatch fires an event to all matching subscriptions.
// Fully async - the API response never waits for delivery.
func Dispatch(event string, data interface{}) {
go func() {
hooks, err := database.ListActiveWebhooksForEvent(event)
if err != nil {
log.Printf("[webhooks] lookup failed: %v", err)
return
}
if len(hooks) == 0 {
return
}
payload := Payload{
Event: event,
Time:  time.Now().UTC().Format(time.RFC3339),
Data:  data,
}
body, err := json.Marshal(payload)
if err != nil {
return
}
for _, h := range hooks {
go deliver(h, event, body)
}
}()
}

// DeliverOne sends one event to one webhook (used by /test endpoint)
func DeliverOne(h database.Webhook, event string, data interface{}) {
payload := Payload{
Event: event,
Time:  time.Now().UTC().Format(time.RFC3339),
Data:  data,
}
body, _ := json.Marshal(payload)
go deliver(h, event, body)
}

func deliver(h database.Webhook, event string, body []byte) {
sig := sign(body, h.Secret)
backoff := time.Second
for attempt := 1; attempt <= 3; attempt++ {
req, err := http.NewRequest("POST", h.URL, bytes.NewReader(body))
if err != nil {
log.Printf("[webhooks] bad url %s: %v", h.URL, err)
return
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("User-Agent", "nexrouter-webhooks/1.0")
req.Header.Set("X-Webhook-Event", event)
req.Header.Set("X-Webhook-Signature", "sha256="+sig)
req.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", attempt))

resp, err := client.Do(req)
if err == nil {
code := resp.StatusCode
resp.Body.Close()
if code >= 200 && code < 300 {
log.Printf("[webhooks] OK %s -> %s (attempt %d, status %d)", event, h.URL, attempt, code)
return
}
log.Printf("[webhooks] %s -> %s status %d (attempt %d)", event, h.URL, code, attempt)
} else {
log.Printf("[webhooks] %s -> %s error: %v (attempt %d)", event, h.URL, err, attempt)
}
if attempt < 3 {
time.Sleep(backoff)
backoff = backoff * 5
}
}
log.Printf("[webhooks] GAVE UP %s -> %s after 3 attempts", event, h.URL)
}

// sign computes HMAC-SHA256 of body with secret (hex)
func sign(body []byte, secret string) string {
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(body)
return hex.EncodeToString(mac.Sum(nil))
}