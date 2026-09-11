package core

import "strings"

type node struct {
path     string
children []*node
handlers []HandlerFunc
}

func (n *node) addRoute(path string, handlers []HandlerFunc) {
segments := splitPath(path)
current := n
for _, seg := range segments {
found := false
for _, child := range current.children {
if child.path == seg {
current = child
found = true
break
}
}
if !found {
child := &node{path: seg}
current.children = append(current.children, child)
current = child
}
}
current.handlers = handlers
}

func (n *node) getValue(path string) ([]HandlerFunc, map[string]string) {
params := make(map[string]string)
segments := splitPath(path)
current := n
for _, seg := range segments {
found := false
for _, child := range current.children {
if child.path == seg {
current = child
found = true
break
}
}
if found {
continue
}
for _, child := range current.children {
if len(child.path) > 0 && child.path[0] == ':' {
params[child.path[1:]] = seg
current = child
found = true
break
}
}
if !found {
return nil, nil
}
}
return current.handlers, params
}

func splitPath(path string) []string {
path = strings.Trim(path, "/")
if path == "" {
return []string{}
}
return strings.Split(path, "/")
}