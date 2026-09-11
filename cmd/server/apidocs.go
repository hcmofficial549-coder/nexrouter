package main

import (
"net/http"

"github.com/nexrouter/nexrouter/core"
"github.com/nexrouter/nexrouter/docs"
)

const docsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>nexrouter API Documentation</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
<style>
body{margin:0;background:#fafafa}
.swagger-ui{max-width:1200px;margin:0 auto}
.topbar{display:none}
.nr-banner{background:linear-gradient(135deg,#3b82f6,#a855f7);color:#fff;padding:18px 32px;font-family:sans-serif}
.nr-banner h1{margin:0;font-size:20px}
.nr-banner p{margin:4px 0 0;font-size:13px;opacity:.85}
</style>
</head>
<body>
<div class="nr-banner">
<h1>nexrouter API v1.7.0</h1>
<p>Custom Go HTTP router - JWT auth, SQLite, caching, realtime SSE. Demo login: john@example.com / password123</p>
</div>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/openapi.json",
    dom_id: "#swagger-ui",
    docExpansion: "list",
    defaultModelsExpandDepth: 1,
    deepLinking: true
  });
};
</script>
</body>
</html>
`

// openAPISpecHandler serves the OpenAPI 3.0 JSON spec
func openAPISpecHandler(c *core.Context) {
c.SetHeader("Content-Type", "application/json; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.SetHeader("Access-Control-Allow-Origin", "*")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(docs.Spec()))
}

// docsHandler serves the interactive Swagger UI page
func docsHandler(c *core.Context) {
c.SetHeader("Content-Type", "text/html; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(docsHTML))
}