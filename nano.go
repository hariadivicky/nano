// Copyright 2020 Vicky Hariadi Pratama. All rights reserved.
// license that can be found in the LICENSE file.
// this package is http route multiplexing.

package nano

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// Middleware struct
type Middleware = func(http.HandlerFunc) http.HandlerFunc

// Route struct
type Route struct {
	Method      string           // current supported method is GET, POST, PUT, and DELETE
	Path        string           // route parameter should having : prefix
	Handler     http.HandlerFunc // should use http.HandlerFunc type.
	Middlewares []Middleware
}

// Router struct
type Router struct {
	Base            string       // base URL must having prefix "/"
	Routes          []*Route     // store route list.
	Middlewares     []Middleware // store router level middleware that will be applied to its route list.
	Groups          []*Router    // the sub router(s).
	notFoundHandler http.HandlerFunc
}

// ParameterBag store current request route parameter
type ParameterBag struct {
	Bag map[string]string // Store router parameter
}

// MatchRoute store route and request parameter to matching request.
type MatchRoute struct {
	ParameterBag
	*Route
}

// Get returns route parameter by given parameter name.
func (pb ParameterBag) Get(name string) string {
	return pb.Bag[name]
}

// Set value to parameter bag.
// it is possible to replace value if given name is already exists.
func (pb ParameterBag) Set(name, value string) {
	pb.Bag[name] = value
}

// Use is to apply middleware function to all Router.Routes
// this function is used to register middleware.
func (router *Router) Use(middleware Middleware) {
	router.Middlewares = append(router.Middlewares, middleware)
}

// Use is to apply middleware function to current route only.
func (route *Route) Use(middleware ...Middleware) {
	route.Middlewares = append(route.Middlewares, middleware...)
}

// Match route to current request.
func (route *Route) Match(r *http.Request) (*MatchRoute, error) {
	matchRoute := new(MatchRoute)

	// there is no matching route since method does not match.
	if route.Method != r.Method {
		return matchRoute, nil
	}

	routeParts := strings.Split(route.Path, "/")
	urlParts := strings.Split(r.URL.Path, "/")

	patterns := make([]string, 0)

	// scanning route parameter placeholder to replace it with regex pattern.
	for key := range routeParts {
		// params path should having ":" prefix
		// e.g. /profile/:id
		if strings.HasPrefix(routeParts[key], ":") {
			patterns = append(patterns, "([a-z0-9_-]+)")
		} else {
			// it is not route parameter placeholder, then store the path part.
			patterns = append(patterns, routeParts[key])
		}
	}

	// create full matching pattern.
	pattern := "^" + strings.Join(patterns, "/") + "$"
	expr, err := regexp.Compile(pattern)

	if err != nil {
		return matchRoute, fmt.Errorf("Cannot compile pattern: %v", err)
	}

	normalizeRequestUrl := r.URL.Path

	// remove "/" suffix from current request url if is not root request "/"
	if r.URL.Path != "/" {
		normalizeRequestUrl = strings.TrimSuffix(r.URL.Path, "/")
	}

	match := expr.Match([]byte(normalizeRequestUrl))

	if !match {
		// current request does not match with current route pattern.
		return matchRoute, nil
	}

	params := make(map[string]string)

	// scanning route parameter and fill it with real value in current request path part.
	for key := range routeParts {
		// params path should having ":" prefix
		// e.g. /profile/:id
		if strings.HasPrefix(routeParts[key], ":") {
			// removing ":" on param
			paramKey := strings.Replace(routeParts[key], ":", "", 1)
			params[paramKey] = urlParts[key]
		}
	}

	matchRoute.ParameterBag = ParameterBag{params}
	matchRoute.Route = route

	return matchRoute, nil
}

// New create new router struct.
func New(args ...string) *Router {
	// default router base URL is "/"
	base := "/"

	if len(args) > 0 {
		// Only accept first arguments as router base URL
		base = args[0]
	}

	router := new(Router)
	router.Base = base

	return router
}

// Group is uses to make sub router group
func (router *Router) Group(prefix string) *Router {
	if router.Base != "/" {
		// Attach parent base URL to sub router prefix.
		prefix = router.Base + prefix
	}

	routerGroup := New(prefix)
	router.Groups = append(router.Groups, routerGroup)

	return routerGroup
}

// Get is to register GET request method.
func (router *Router) Get(path string, handler http.HandlerFunc) *Route {
	return router.PushRoute(http.MethodGet, path, handler)
}

// Post is to register POST request method.
func (router *Router) Post(path string, handler http.HandlerFunc) *Route {
	return router.PushRoute(http.MethodPost, path, handler)
}

// Put is to register PUT request method.
func (router *Router) Put(path string, handler http.HandlerFunc) *Route {
	return router.PushRoute(http.MethodPut, path, handler)
}

// Delete is to register DELETE request method.
func (router *Router) Delete(path string, handler http.HandlerFunc) *Route {
	return router.PushRoute(http.MethodDelete, path, handler)
}

// PushRoute route into main routes list.
func (router *Router) PushRoute(method, path string, handler http.HandlerFunc) *Route {

	// only add path with current router base URL when the base URL is not "/".
	// this also remove "/" suffix
	if router.Base != "/" {
		path = strings.TrimSuffix(router.Base+path, "/")
	}

	route := &Route{method, path, handler, nil}

	if len(router.Middlewares) > 0 {
		// Apply router level middleware to current route.
		route.Use(router.Middlewares...)
	}

	router.Routes = append(router.Routes, route)

	return route
}

// ParseParameter is to parse matched parameter pattern in current request which stored in ParameterBag.
func (router *Router) ParseParameter(r *http.Request) ParameterBag {
	ctx := r.Context()
	return ctx.Value("params").(ParameterBag)
}

// Match is used to find matches route in router to current request
func (router *Router) Match(r *http.Request) *MatchRoute {
	var found *MatchRoute

	for _, route := range router.Routes {
		match, err := route.Match(r)

		// Matching error, stop matching process.
		if err != nil {
			log.Printf("error while matching route: %v \n", err)
			break
		}

		if match.Route != nil {
			// stop matching route when found matches.
			found = match
			break
		}
	}

	return found
}

// DefaultNotFoundHandler called when custom not found handler is not set.
func (router *Router) DefaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found.\n"))
}

func (router *Router) NotFoundHandler(handler http.HandlerFunc) {
	router.notFoundHandler = handler
}

// ServeHTTP implementation.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var matches *MatchRoute

	// Matching main router.
	wg.Add(1)
	go func() {
		matches = router.Match(r)
		wg.Done()
	}()

	wg.Wait()

	if matches == nil && r.URL.Path != "/" {
		// There is no matching route in main router, then find it in sub router
		urlParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		baseRequest := "/" + urlParts[0]

		if len(router.Groups) > 0 {

			wg.Add(len(router.Groups))
			// try to find current request to sub router.
			for key := range router.Groups {
				// find routes if current router group base URL is match with current request base URL.
				go func(routerGroup *Router) {
					if routerGroup.Base == baseRequest {
						matches = routerGroup.Match(r)
					}
					wg.Done()
				}(router.Groups[key])
			}

			wg.Wait()
		}
	}

	// no matching route, show not found.
	if matches == nil {
		if router.notFoundHandler != nil {
			router.notFoundHandler(w, r)
			return
		}
		// use default not found handler when custom not found handler is not set.
		router.DefaultNotFoundHandler(w, r)
		return
	}

	// Below is flow for found match route.
	//
	// # a happy duck
	// <(' )____//
	//  \______/
	//   /   \

	// add route parameter bag to current request context.
	ctx := context.WithValue(context.Background(), "params", matches.ParameterBag)
	r = r.WithContext(ctx)

	handler := matches.Route.Handler

	// check route middleware.
	if len(matches.Middlewares) > 0 {
		// there are exists route middlewares, apply it to handler.
		if len(matches.Middlewares) == 1 {
			// is this efficient, LOL
			handler = matches.Middlewares[0](handler)
		} else {
			// reverse middleware stack.
			for i := len(matches.Middlewares) - 1; i >= 0; i-- {
				handler = matches.Middlewares[i](handler)
			}
		}
	}

	// forwarding request to route handler.
	handler(w, r)
}
