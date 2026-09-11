package core

import (
"encoding/json"
"io"
"net/http"
"strings"
)

// Context wraps http.Request and http.ResponseWriter
type Context struct {
Request  *http.Request
Writer   http.ResponseWriter
handlers []HandlerFunc
index    int
params   map[string]string
keys     map[string]interface{}
status   int
written  bool
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
c.Writer = w
c.Request = r
c.index = -1
c.status = http.StatusOK
c.written = false
for k := range c.params {
delete(c.params, k)
}
for k := range c.keys {
delete(c.keys, k)
}
}

// Next executes the next handler in the chain
func (c *Context) Next() {
c.index++
for c.index < len(c.handlers) {
c.handlers[c.index](c)
c.index++
}
}

// Abort stops the middleware chain
func (c *Context) Abort() { c.index = len(c.handlers) }

// AbortWithStatus aborts and writes a status code
func (c *Context) AbortWithStatus(code int) {
c.Writer.WriteHeader(code)
c.status = code
c.written = true
c.Abort()
}

// StatusCode returns the response status code
func (c *Context) StatusCode() int { return c.status }

// Written reports whether a response has been written
func (c *Context) Written() bool { return c.written }

// Param returns a URL parameter
func (c *Context) Param(key string) string { return c.params[key] }

// Query returns a query string parameter
func (c *Context) Query(key string) string { return c.Request.URL.Query().Get(key) }

// GetHeader returns a request header
func (c *Context) GetHeader(key string) string { return c.Request.Header.Get(key) }

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) { c.Writer.Header().Set(key, value) }

// Status sets the response status code
func (c *Context) Status(code int) { c.status = code }

// JSON writes a JSON response
func (c *Context) JSON(code int, obj interface{}) {
c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
c.Writer.WriteHeader(code)
c.status = code
c.written = true
json.NewEncoder(c.Writer).Encode(obj)
}

// String writes a plain text response
func (c *Context) String(code int, s string) {
c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
c.Writer.WriteHeader(code)
c.status = code
c.written = true
c.Writer.Write([]byte(s))
}

// OK writes a 200 JSON response
func (c *Context) OK(obj interface{}) { c.JSON(http.StatusOK, obj) }

// Created writes a 201 JSON response
func (c *Context) Created(obj interface{}) { c.JSON(http.StatusCreated, obj) }

// BadRequest writes a 400 JSON error
func (c *Context) BadRequest(msg string) { c.JSON(http.StatusBadRequest, H{"error": msg}) }

// Unauthorized writes a 401 JSON error
func (c *Context) Unauthorized(msg string) { c.JSON(http.StatusUnauthorized, H{"error": msg}) }

// NotFound writes a 404 JSON error
func (c *Context) NotFound(msg string) { c.JSON(http.StatusNotFound, H{"error": msg}) }

// InternalError writes a 500 JSON error
func (c *Context) InternalError(msg string) {
c.JSON(http.StatusInternalServerError, H{"error": msg})
}

// BindJSON decodes the request body into obj
func (c *Context) BindJSON(obj interface{}) error {
body, err := io.ReadAll(c.Request.Body)
if err != nil {
return err
}
defer c.Request.Body.Close()
return json.Unmarshal(body, obj)
}

// Set stores a value in the context
func (c *Context) Set(key string, value interface{}) { c.keys[key] = value }

// Get retrieves a value from the context
func (c *Context) Get(key string) (interface{}, bool) {
val, exists := c.keys[key]
return val, exists
}

// ClientIP returns the best-guess client IP
func (c *Context) ClientIP() string {
if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
return strings.Split(ip, ",")[0]
}
if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
return ip
}
ip := c.Request.RemoteAddr
if i := strings.LastIndex(ip, ":"); i != -1 {
ip = ip[:i]
}
return ip
}