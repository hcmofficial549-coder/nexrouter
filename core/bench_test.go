package core

import (
"net/http"
"net/http/httptest"
"testing"
)

// nopWriter avoids ResponseRecorder memory growth in benchmarks
type nopWriter struct {
h    http.Header
code int
}

func newNopWriter() *nopWriter {
return &nopWriter{h: make(http.Header)}
}

func (w *nopWriter) Header() http.Header         { return w.h }
func (w *nopWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *nopWriter) WriteHeader(code int)        { w.code = code }

func setupBenchRouter() *Router {
r := New()
r.GET("/health", func(c *Context) { c.String(200, "ok") })
r.GET("/api/v1/products", func(c *Context) { c.JSON(200, H{"ok": true}) })
r.GET("/api/v1/products/:id", func(c *Context) { c.JSON(200, H{"id": c.Param("id")}) })
r.GET("/api/v1/users/:id/posts/:postID", func(c *Context) { c.JSON(200, H{}) })
r.POST("/api/v1/products", func(c *Context) { c.JSON(201, H{}) })
return r
}

func BenchmarkRouteStatic(b *testing.B) {
r := setupBenchRouter()
w := newNopWriter()
req := httptest.NewRequest("GET", "/health", nil)
b.ReportAllocs()
b.ResetTimer()
for i := 0; i < b.N; i++ {
r.ServeHTTP(w, req)
}
}

func BenchmarkRouteParam(b *testing.B) {
r := setupBenchRouter()
w := newNopWriter()
req := httptest.NewRequest("GET", "/api/v1/products/42", nil)
b.ReportAllocs()
b.ResetTimer()
for i := 0; i < b.N; i++ {
r.ServeHTTP(w, req)
}
}

func BenchmarkRouteMultiParam(b *testing.B) {
r := setupBenchRouter()
w := newNopWriter()
req := httptest.NewRequest("GET", "/api/v1/users/7/posts/99", nil)
b.ReportAllocs()
b.ResetTimer()
for i := 0; i < b.N; i++ {
r.ServeHTTP(w, req)
}
}

func BenchmarkRouteNotFound(b *testing.B) {
r := setupBenchRouter()
w := newNopWriter()
req := httptest.NewRequest("GET", "/tidak-ada", nil)
b.ReportAllocs()
b.ResetTimer()
for i := 0; i < b.N; i++ {
r.ServeHTTP(w, req)
}
}