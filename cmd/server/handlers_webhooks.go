package main

import (
"crypto/rand"
"encoding/hex"
"net/http"
"strconv"
"strings"

"github.com/nexrouter/nexrouter/core"
"github.com/nexrouter/nexrouter/database"
"github.com/nexrouter/nexrouter/webhooks"
)

// generateSecret makes a random hex secret (24 bytes)
func generateSecret() string {
b := make([]byte, 24)
if _, err := rand.Read(b); err != nil {
return "nexrouter-fallback-secret"
}
return hex.EncodeToString(b)
}

// createWebhookHandler POST /api/v1/webhooks (protected)
func createWebhookHandler(c *core.Context) {
var input struct {
URL    string   `json:"url"`
Events []string `json:"events"`
Secret string   `json:"secret"`
}
if err := c.BindJSON(&input); err != nil {
c.BadRequest("invalid request body")
return
}
if !strings.HasPrefix(input.URL, "http://") && !strings.HasPrefix(input.URL, "https://") {
c.BadRequest("url must start with http:// or https://")
return
}
events := "*"
if len(input.Events) > 0 {
events = strings.Join(input.Events, ",")
}
secret := input.Secret
if secret == "" {
secret = generateSecret()
}
w, err := database.CreateWebhook(input.URL, events, secret)
if err != nil {
c.InternalError(err.Error())
return
}
c.Created(core.H{
"success": true,
"data":    w,
"hint":    "simpan secret ini - dipakai verifikasi header X-Webhook-Signature (HMAC-SHA256 hex dari body)",
})
}

// listWebhooksHandler GET /api/v1/webhooks (protected)
func listWebhooksHandler(c *core.Context) {
items, err := database.ListWebhooks()
if err != nil {
c.InternalError(err.Error())
return
}
c.JSON(http.StatusOK, core.H{"success": true, "data": items, "count": len(items)})
}

// deleteWebhookHandler DELETE /api/v1/webhooks/:id (protected)
func deleteWebhookHandler(c *core.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.BadRequest("invalid webhook id")
return
}
ok, err := database.DeleteWebhook(id)
if err != nil {
c.InternalError(err.Error())
return
}
if !ok {
c.NotFound("webhook not found")
return
}
c.JSON(http.StatusOK, core.H{"success": true, "message": "webhook deleted"})
}

// testWebhookHandler POST /api/v1/webhooks/:id/test (protected)
func testWebhookHandler(c *core.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.BadRequest("invalid webhook id")
return
}
w, err := database.GetWebhook(id)
if err != nil {
c.InternalError(err.Error())
return
}
if w == nil {
c.NotFound("webhook not found")
return
}
webhooks.DeliverOne(*w, "webhook.test", core.H{
"message":    "Hello from nexrouter webhooks!",
"webhook_id": w.ID,
})
c.JSON(http.StatusOK, core.H{
"success": true,
"message": "test event dispatched async - cek receiver & docker logs",
})
}