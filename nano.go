// Copyright 2020 Vicky Hariadi Pratama. All rights reserved.
// license that can be found in the LICENSE file.
// this package is http route multiplexing.

package nano

import (
	"log"
	"net/http"
	"strings"
)

const (
	// HeaderAcceptEncoding is accept encoding.
	HeaderAcceptEncoding = "Accept-Encoding"
	// HeaderContentEncoding is content encoding.
	HeaderContentEncoding = "Content-Encoding"
	// HeaderContentLength is content length.
	HeaderContentLength = "Content-Length"
	// HeaderContentType is content type.
	HeaderContentType = "Content-Type"
	// HeaderAccept is accept content type.
	HeaderAccept = "Accept"
	// HeaderOrigin is request origin.
	HeaderOrigin = "Origin"
	// HeaderVary is request vary.
	HeaderVary = "Vary"
	// HeaderAccessControlRequestMethod is cors request method.
	HeaderAccessControlRequestMethod = "Access-Control-Request-Method"
	// HeaderAccessControlRequestHeader is cors request header.
	HeaderAccessControlRequestHeader = "Access-Control-Request-Header"
	// HeaderAccessControlAllowOrigin is cors allowed origins.
	HeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
	// HeaderAccessControlAllowMethods is cors allowed origins.
	HeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"
	// HeaderAccessControlAllowHeader is cors allowed headers.
	HeaderAccessControlAllowHeader = "Access-Control-Allow-Header"

	// MimeJSON is standard json mime.
	MimeJSON = "application/json"
	// MimeHTML is standard html mime.
	MimeHTML = "text/html"
	// MimePlainText is standard plain text mime.
	MimePlainText = "text/plain"
)

// Engine defines nano web engine.
type Engine struct {
	*RouterGroup
	router *router
	debug  bool
	groups []*RouterGroup
}

// RouterGroup defines collection of route that has same prefix
type RouterGroup struct {
	prefix      string
	engine      *Engine
	middlewares []HandlerFunc
	parent      *RouterGroup
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

// Use is functions to apply middleware function(s).
func (rg *RouterGroup) Use(middlewares ...HandlerFunc) {
	rg.middlewares = append(rg.middlewares, middlewares...)
}

// Group is functions to create new router group.
func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	group := &RouterGroup{
		prefix: rg.prefix + prefix,
		parent: rg,
		engine: rg.engine,
	}

	rg.engine.groups = append(rg.engine.groups, group)

	return group
}

// GET is functions to register route with GET request method.
func (rg *RouterGroup) GET(urlPattern string, handler ...HandlerFunc) {
	rg.addRoute(http.MethodGet, urlPattern, handler...)
}

// POST is functions to register route with POST request method.
func (rg *RouterGroup) POST(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodPost, urlPattern, handler)
}

// PUT is functions to register route with PUT request method.
func (rg *RouterGroup) PUT(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodPut, urlPattern, handler)
}

// DELETE is functions to register route with DELETE request method.
func (rg *RouterGroup) DELETE(urlPattern string, handler HandlerFunc) {
	rg.addRoute(http.MethodDelete, urlPattern, handler)
}

// Default is functions to register default handler when no matching routes.
// Only one Default handler allowed to register.
func (rg *RouterGroup) Default(handler HandlerFunc) {
	// reject overriding.
	if rg.engine.router.defaultHandler != nil {
		log.Fatal("could not register default handler because it already registered\n")
	}

	rg.engine.router.defaultHandler = handler
}

// addRoute is functions to register new route with current group prefix.
func (rg *RouterGroup) addRoute(requestMethod, urlPattern string, handler ...HandlerFunc) {
	// append router group prefix.
	prefixedURLPattern := rg.prefix + urlPattern

	rg.engine.router.addRoute(requestMethod, prefixedURLPattern, handler...)
}

// ServeHTTP implements multiplexer.
func (ng *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middlewares := make([]HandlerFunc, 0)

	// scanning for router group middleware.
	for _, group := range ng.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	ctx := newContext(w, r)
	ctx.handlers = middlewares
	ng.router.handle(ctx)
}

// Run applications.
func (ng *Engine) Run(address string) error {
	return http.ListenAndServe(address, ng)
}
