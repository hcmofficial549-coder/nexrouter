package main

import (
"net/http"

"github.com/nexrouter/nexrouter/core"
"github.com/nexrouter/nexrouter/docs"
)

const docsHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport"
content="width=device-width, initial-scale=1">
<title>nexrouter API Docs</title>
<link rel="stylesheet"
href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script
src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js">
</script><script>
window.onload = function() {
  SwaggerUIBundle({
    url: "/openapi.json",
    dom_id: "#swagger-ui",
    docExpansion: "list",
    deepLinking: true
  });
};
</script>
</body>
</html>
`

func openAPISpecHandler(c *core.Context) {
c.SetHeader("Content-Type", "application/json; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(docs.Spec()))
}

func docsHandler(c *core.Context) {
c.SetHeader("Content-Type", "text/html; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(docsHTML))
}