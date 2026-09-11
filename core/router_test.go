package core

import (
"net/http/httptest"
"testing"
)

func TestRouterGET(t *testing.T) {
r := New()
called := false
r.GET("/test", func(c *Context) {
called = true
c.JSON(200, H{"message": "ok"})
})
req := httptest.NewRequest("GET", "/test", nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
if !called {
t.Error("handler not called")
}
if w.Code != 200 {
t.Errorf("expected 200, got %d", w.Code)
}
}

func TestRouterParam(t *testing.T) {
r := New()
r.GET("/users/:id", func(c *Context) {
c.JSON(200, H{"id": c.Param("id")})
})
req := httptest.NewRequest("GET", "/users/42", nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
if w.Code != 200 {
t.Errorf("expected 200, got %d", w.Code)
}
if !contains(w.Body.String(), "42") {
t.Errorf("body should contain 42: %s", w.Body.String())
}
}

func TestRouter404(t *testing.T) {
r := New()
req := httptest.NewRequest("GET", "/nothing", nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
if w.Code != 404 {
t.Errorf("expected 404, got %d", w.Code)
}
}

func TestRouteGroup(t *testing.T) {
r := New()
api := r.Group("/api/v1")
api.GET("/ping", func(c *Context) { c.JSON(200, H{"pong": true}) })
req := httptest.NewRequest("GET", "/api/v1/ping", nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
if w.Code != 200 {
t.Errorf("expected 200, got %d", w.Code)
}
}

func TestMiddlewareOrder(t *testing.T) {
r := New()
var order []int
r.Use(func(c *Context) { order = append(order, 1); c.Next(); order = append(order, 4) })
r.Use(func(c *Context) { order = append(order, 2); c.Next(); order = append(order, 3) })
r.GET("/x", func(c *Context) { c.String(200, "x") })
req := httptest.NewRequest("GET", "/x", nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
if len(order) != 4 || order[0] != 1 || order[1] != 2 || order[2] != 3 || order[3] != 4 {
t.Errorf("wrong middleware order: %v", order)
}
}

func contains(s, sub string) bool {
return len(s) >= len(sub) && searchStr(s, sub)
}

func searchStr(s, sub string) bool {
for i := 0; i+len(sub) <= len(s); i++ {
if s[i:i+len(sub)] == sub {
return true
}
}
return false
}