package nano

import (
	"fmt"
	"net/http"
	"strings"
)

// router defines main router structure.
type router struct {
	nodes          map[string]*node
	handlers       map[string][]HandlerFunc
	defaultHandler HandlerFunc
}

// newRouter is router constructor.
func newRouter() *router {
	return &router{
		nodes:    make(map[string]*node),
		handlers: make(map[string][]HandlerFunc),
	}
}

// createUrlParts returns splitted path.
func createURLParts(urlPattern string) []string {
	patternParts := strings.Split(urlPattern, "/")

	urlParts := make([]string, 0)

	for _, path := range patternParts {
		// ignore root path
		if path != "" {
			urlParts = append(urlParts, path)

			// only * wildcard is allowed.
			if path[0] == '*' {
				break
			}
		}
	}

	return urlParts
}

// addRoute functions to register route to router.
func (r *router) addRoute(requestMethod, urlPattern string, handler ...HandlerFunc) {
	urlParts := createURLParts(urlPattern)

	rootNode, exists := r.nodes[requestMethod]

	// current request method root node doesn't exists.
	if !exists {
		r.nodes[requestMethod] = &node{}
		rootNode = r.nodes[requestMethod]
	}

	// register route.
	key := fmt.Sprintf("%s-%s", requestMethod, urlPattern)

	// insert children to tree.
	rootNode.insertChildren(urlPattern, urlParts, 0)
	r.handlers[key] = handler
}

// findRoute functions to matching current request with route node.
func (r *router) findRoute(requestMethod, urlPath string) (*node, map[string]string) {
	searchParts := createURLParts(urlPath)
	params := make(map[string]string)

	rootNode, exists := r.nodes[requestMethod]

	// there are no routes with current request method
	if !exists {
		return nil, nil
	}

	// scan child node recursively.
	node := rootNode.findNode(searchParts, 0)

	if node != nil {
		// replace param placeholder with current request value.
		for index, path := range createURLParts(node.urlPattern) {
			// current pattern is parameter.
			if path[0] == ':' {
				params[path[1:]] = searchParts[index]
			}

			// current pattern is * wildcard, that means all path are used.
			if path[0] == '*' && len(path) > 1 {
				params[path[1:]] = strings.Join(searchParts[index:], "/")
			}
		}

		return node, params
	}

	return nil, nil
}

// notFoundHandler is router default handler.
func (r *router) notFoundHandler() HandlerFunc {
	return func(c *Context) {
		c.String(http.StatusOK, "Nano - 404 %s not found", c.Path)
	}
}

// serveDefaultHandler will append default handler to call stacks.
func (r *router) serveDefaultHandler(c *Context) {
	// create not found handler when default handler not set yet.
	if r.defaultHandler == nil {
		r.defaultHandler = r.notFoundHandler()
	}

	c.handlers = append(c.handlers, r.defaultHandler)
	c.Next()
}

// handle incoming request.
func (r *router) handle(c *Context) {
	node, params := r.findRoute(c.Method, c.Path)

	// current request has a match route.
	if node != nil {
		key := fmt.Sprintf("%s-%s", c.Method, node.urlPattern)
		c.Params = params

		// append current handler to handler stack.
		// extract route handler(s).
		c.handlers = append(c.handlers, r.handlers[key]...)
	} else {
		// no matching routes, serve default.
		r.serveDefaultHandler(c)
	}

	// call handlers stack.
	c.Next()
}
