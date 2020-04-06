// Copyright 2020 Vicky Hariadi Pratama. All rights reserved.
// license that can be found in the LICENSE file.
// this package is http route multiplexing.

package nano

import "net/http"

// Engine defines nano web engine.
type Engine struct {
	router *router
	debug  bool
}

// H defines json wrapper.
type H map[string]interface{}

// HandlerFunc defines nano request handler function signature.
type HandlerFunc func(c *Context)

// New is nano constructor
func New() *Engine {
	return &Engine{
		router: newRouter(),
		debug:  false,
	}
}

// GET functions to register route with GET request method.
func (ng *Engine) GET(urlPattern string, handler HandlerFunc) {
	ng.router.addRoute(http.MethodGet, urlPattern, handler)
}

// POST functions to register route with POST request method.
func (ng *Engine) POST(urlPattern string, handler HandlerFunc) {
	ng.router.addRoute(http.MethodPost, urlPattern, handler)
}

// PUT functions to register route with PUT request method.
func (ng *Engine) PUT(urlPattern string, handler HandlerFunc) {
	ng.router.addRoute(http.MethodPut, urlPattern, handler)
}

// DELETE functions to register route with DELETE request method.
func (ng *Engine) DELETE(urlPattern string, handler HandlerFunc) {
	ng.router.addRoute(http.MethodDelete, urlPattern, handler)
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
