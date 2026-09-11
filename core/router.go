package core

import (
"net/http"
"strings"
"sync"
)

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

// HandlerFunc defines the handler function signature
type HandlerFunc func(*Context)

// Router is the main HTTP router
type Router struct {
trees      map[string]*node
pool       sync.Pool
middleware []HandlerFunc
maxParams  uint16
}

// New creates a new Router
func New() *Router {
r := &Router{trees: make(map[string]*node)}
r.pool.New = func() interface{} {
return &Context{
params: make(map[string]string),
keys:   make(map[string]interface{}),
}
}
return r
}

// Use adds global middleware
func (r *Router) Use(mw ...HandlerFunc) {
r.middleware = append(r.middleware, mw...)
}

// GET registers a GET route
func (r *Router) GET(path string, h ...HandlerFunc) { r.addRoute(http.MethodGet, path, h) }

// POST registers a POST route
func (r *Router) POST(path string, h ...HandlerFunc) { r.addRoute(http.MethodPost, path, h) }

// PUT registers a PUT route
func (r *Router) PUT(path string, h ...HandlerFunc) { r.addRoute(http.MethodPut, path, h) }

// DELETE registers a DELETE route
func (r *Router) DELETE(path string, h ...HandlerFunc) { r.addRoute(http.MethodDelete, path, h) }

// PATCH registers a PATCH route
func (r *Router) PATCH(path string, h ...HandlerFunc) { r.addRoute(http.MethodPatch, path, h) }

// OPTIONS registers an OPTIONS route
func (r *Router) OPTIONS(path string, h ...HandlerFunc) { r.addRoute(http.MethodOptions, path, h) }

// Group creates a route group with shared prefix
func (r *Router) Group(prefix string, h ...HandlerFunc) *Group {
return &Group{prefix: prefix, router: r, middleware: h}
}

func (r *Router) addRoute(method, path string, handlers []HandlerFunc) {
if path == "" || path[0] != '/' {
panic("nexrouter: path must begin with '/'")
}
root := r.trees[method]
if root == nil {
root = new(node)
r.trees[method] = root
}
root.addRoute(path, handlers)
n := strings.Count(path, ":") + strings.Count(path, "*")
if uint16(n) > r.maxParams {
r.maxParams = uint16(n)
}
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
return http.ListenAndServe(addr, r)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
c := r.pool.Get().(*Context)
c.reset(w, req)
root := r.trees[req.Method]
if root != nil {
handlers, params := root.getValue(req.URL.Path)
if handlers != nil {
c.params = params
c.handlers = make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
c.handlers = append(c.handlers, r.middleware...)
c.handlers = append(c.handlers, handlers...)
c.Next()
r.pool.Put(c)
return
}
}
c.JSON(http.StatusNotFound, H{"error": "404 page not found", "path": req.URL.Path})
r.pool.Put(c)
}

// Group represents a route group
type Group struct {
prefix     string
router     *Router
middleware []HandlerFunc
}

// Use adds middleware to the group
func (g *Group) Use(mw ...HandlerFunc) {
g.middleware = append(g.middleware, mw...)
}

func (g *Group) addRoute(method, path string, handlers []HandlerFunc) {
combined := make([]HandlerFunc, 0, len(g.middleware)+len(handlers))
combined = append(combined, g.middleware...)
combined = append(combined, handlers...)
g.router.addRoute(method, g.prefix+path, combined)
}

// GET registers a GET route in the group
func (g *Group) GET(path string, h ...HandlerFunc) { g.addRoute(http.MethodGet, path, h) }

// POST registers a POST route in the group
func (g *Group) POST(path string, h ...HandlerFunc) { g.addRoute(http.MethodPost, path, h) }

// PUT registers a PUT route in the group
func (g *Group) PUT(path string, h ...HandlerFunc) { g.addRoute(http.MethodPut, path, h) }

// DELETE registers a DELETE route in the group
func (g *Group) DELETE(path string, h ...HandlerFunc) { g.addRoute(http.MethodDelete, path, h) }