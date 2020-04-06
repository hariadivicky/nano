// Copyright 2020 Vicky Hariadi Pratama. All rights reserved.
// license that can be found in the LICENSE file.
// this package is http route multiplexing.

package nano

import "net/http"

// RouterGroup defines collection of route that has same prefix
type RouterGroup struct {
	prefix string
	engine *Engine
	parent *RouterGroup
}

// Engine defines nano web engine.
type Engine struct {
	*RouterGroup
	router *router
	debug  bool
	groups []*RouterGroup
}

// H defines json wrapper.
type H map[string]interface{}

// HandlerFunc defines nano request handler function signature.
type HandlerFunc func(c *Context)

// New is nano constructor
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
		debug:  false,
	}

	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}

	return engine
}

// Group functions to create new router group.
func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	group := &RouterGroup{
		prefix: rg.prefix + prefix,
		parent: rg,
		engine: rg.engine,
	}

	rg.engine.groups = append(rg.engine.groups, group)

	return group
}

// GET functions to register route with GET request method.
func (rg *RouterGroup) GET(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodGet, urlPattern, handler)
}

// POST functions to register route with POST request method.
func (rg *RouterGroup) POST(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodPost, urlPattern, handler)
}

// PUT functions to register route with PUT request method.
func (rg *RouterGroup) PUT(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodPut, urlPattern, handler)
}

// DELETE functions to register route with DELETE request method.
func (rg *RouterGroup) DELETE(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodDelete, urlPattern, handler)
}

// addRoute functions to register new route with current group prefix.
func (rg *RouterGroup) addRoute(requestMethod, urlPattern string, handler HandlerFunc) {
	// append router group prefix.
	prefixedURLPattern := rg.prefix + urlPattern

	rg.engine.router.addRoute(requestMethod, prefixedURLPattern, handler)
}

// ServeHTTP implements multiplexer.
func (ng *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r)
	ng.router.handle(ctx)
}

// Run applications.
func (ng *Engine) Run(address string) error {
	return http.ListenAndServe(address, ng)
}
