package main

import (
_ "embed"
"net/http"

"github.com/nexrouter/nexrouter/core"
)

//go:embed dashboard/index.html
var dashboardHTML string

// registerDashboard serves the embedded frontend at /dashboard
func registerDashboard(r *core.Router) {
r.GET("/dashboard", func(c *core.Context) {
c.SetHeader("Content-Type", "text/html; charset=utf-8")
c.SetHeader("Cache-Control", "no-cache")
c.Writer.WriteHeader(http.StatusOK)
c.Writer.Write([]byte(dashboardHTML))
})
}