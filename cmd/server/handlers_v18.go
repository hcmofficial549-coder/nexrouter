package main

import (
"net/http"
"strconv"
"time"

"github.com/nexrouter/nexrouter/core"
"github.com/nexrouter/nexrouter/database"
)

// atoiDefault parses int from query with fallback
func atoiDefault(s string, def int) int {
if s == "" {
return def
}
v, err := strconv.Atoi(s)
if err != nil {
return def
}
return v
}

// listProductsPaged handles GET /api/v1/products
// Query params: page, limit, search, category, sort, order
func listProductsPaged(c *core.Context) {
key := "products:" + c.Request.URL.RawQuery

if data, ok := appCache.Get(key); ok {
c.SetHeader("X-Cache", "HIT")
c.JSON(http.StatusOK, data)
return
}

q := &database.ProductQuery{
Page:     atoiDefault(c.Query("page"), 1),
Limit:    atoiDefault(c.Query("limit"), 50),
Search:   c.Query("search"),
Category: c.Query("category"),
Sort:     c.Query("sort"),
Order:    c.Query("order"),
}

result, err := database.QueryProducts(q)
if err != nil {
c.InternalError(err.Error())
return
}

resp := core.H{
"success": true,
"data":    result.Data,
"pagination": core.H{
"page":        result.Page,
"limit":       result.Limit,
"total":       result.Total,
"total_pages": result.TotalPages,
},
}
appCache.Set(key, resp, 30*time.Second)
c.SetHeader("X-Cache", "MISS")
c.JSON(http.StatusOK, resp)
}

// listUsersPaged handles GET /api/v1/users
// Query params: page, limit, search (name or email)
func listUsersPaged(c *core.Context) {
q := &database.UserQuery{
Page:   atoiDefault(c.Query("page"), 1),
Limit:  atoiDefault(c.Query("limit"), 50),
Search: c.Query("search"),
}
result, err := database.QueryUsers(q)
if err != nil {
c.InternalError(err.Error())
return
}
c.JSON(http.StatusOK, core.H{
"success": true,
"data":    result.Data,
"pagination": core.H{
"page":        result.Page,
"limit":       result.Limit,
"total":       result.Total,
"total_pages": result.TotalPages,
},
})
}