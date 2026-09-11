package database

import (
"fmt"
"strings"
)

// ============ Products: pagination + search + sort ============

// ProductQuery holds list options
type ProductQuery struct {
Page     int
Limit    int
Search   string
Category string
Sort     string
Order    string
}

// ProductPage is a paginated result
type ProductPage struct {
Data       []Product `json:"data"`
Total      int       `json:"total"`
Page       int       `json:"page"`
Limit      int       `json:"limit"`
TotalPages int       `json:"total_pages"`
}

// normalize clamps values and WHITELISTS the sort column
// (never interpolate raw user input into SQL!)
func (q *ProductQuery) normalize() {
if q.Page < 1 {
q.Page = 1
}
if q.Limit < 1 {
q.Limit = 50
}
if q.Limit > 100 {
q.Limit = 100
}
switch q.Sort {
case "name", "price", "stock", "category":
default:
q.Sort = "id"
}
if strings.ToLower(q.Order) == "desc" {
q.Order = "DESC"
} else {
q.Order = "ASC"
}
}

func (q *ProductQuery) where() (string, []interface{}) {
var conds []string
var args []interface{}
if q.Search != "" {
conds = append(conds, "name LIKE ?")
args = append(args, "%"+q.Search+"%")
}
if q.Category != "" {
conds = append(conds, "category = ?")
args = append(args, q.Category)
}
if len(conds) == 0 {
return "", args
}
return " WHERE " + strings.Join(conds, " AND "), args
}

// QueryProducts returns one page of products + totals
func QueryProducts(q *ProductQuery) (*ProductPage, error) {
q.normalize()
w, args := q.where()

var total int
if err := DB.QueryRow("SELECT COUNT(*) FROM products"+w, args...).Scan(&total); err != nil {
return nil, err
}

offset := (q.Page - 1) * q.Limit
sql := fmt.Sprintf(
"SELECT id, name, price, stock, category FROM products%s ORDER BY %s %s LIMIT ? OFFSET ?",
w, q.Sort, q.Order,
)
rows, err := DB.Query(sql, append(args, q.Limit, offset)...)
if err != nil {
return nil, err
}
defer rows.Close()

items := []Product{}
for rows.Next() {
var p Product
if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category); err != nil {
return nil, err
}
items = append(items, p)
}
if err := rows.Err(); err != nil {
return nil, err
}

totalPages := 1
if q.Limit > 0 {
totalPages = (total + q.Limit - 1) / q.Limit
}
return &ProductPage{
Data: items, Total: total,
Page: q.Page, Limit: q.Limit, TotalPages: totalPages,
}, nil
}

// ============ Users: pagination + search ============

// UserQuery holds list options
type UserQuery struct {
Page   int
Limit  int
Search string
}

// UserPage is a paginated result
type UserPage struct {
Data       []User `json:"data"`
Total      int    `json:"total"`
Page       int    `json:"page"`
Limit      int    `json:"limit"`
TotalPages int    `json:"total_pages"`
}

// QueryUsers returns one page of users (search hits name OR email)
func QueryUsers(q *UserQuery) (*UserPage, error) {
if q.Page < 1 {
q.Page = 1
}
if q.Limit < 1 {
q.Limit = 50
}
if q.Limit > 100 {
q.Limit = 100
}

w := ""
var args []interface{}
if q.Search != "" {
w = " WHERE (name LIKE ? OR email LIKE ?)"
args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
}

var total int
if err := DB.QueryRow("SELECT COUNT(*) FROM users"+w, args...).Scan(&total); err != nil {
return nil, err
}

offset := (q.Page - 1) * q.Limit
rows, err := DB.Query("SELECT id, name, email FROM users"+w+" ORDER BY id ASC LIMIT ? OFFSET ?", append(args, q.Limit, offset)...)
if err != nil {
return nil, err
}
defer rows.Close()

items := []User{}
for rows.Next() {
var u User
if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
return nil, err
}
items = append(items, u)
}
if err := rows.Err(); err != nil {
return nil, err
}

totalPages := 1
if q.Limit > 0 {
totalPages = (total + q.Limit - 1) / q.Limit
}
return &UserPage{
Data: items, Total: total,
Page: q.Page, Limit: q.Limit, TotalPages: totalPages,
}, nil
}