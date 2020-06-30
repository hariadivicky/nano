package nano

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateURLParts(t *testing.T) {
	tt := []struct {
		name   string
		url    string
		slices []string
	}{
		{"root url", "/", []string{}},
		{"one url part", "/home", []string{"home"}},
		{"one url part without backslash prefix", "home", []string{"home"}},
		{"one url part with backslash suffix", "home/", []string{"home"}},
		{"two url parts", "/home/services", []string{"home", "services"}},
		{"two url parts without backslash prefix", "home/services", []string{"home", "services"}},
		{"two url parts with backslash suffix", "home/services/", []string{"home", "services"}},
		{"three url parts with star wildcard", "/downloads/*file", []string{"downloads", "*file"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(st *testing.T) {
			rs := createURLParts(tc.url)

			if ln := len(rs); ln != len(tc.slices) {
				st.Errorf("expected result length to be %d; got %d", len(tc.slices), ln)
			}

			if len(tc.slices) > 0 {
				for i, urlPart := range rs {
					if urlPart != tc.slices[i] {
						st.Errorf("expected url part at index %d to be %s; got %s", i, tc.slices[i], urlPart)
					}
				}
			}
		})
	}
}

func TestCreateRoute(t *testing.T) {
	router := newRouter()

	if handlersLen := len(router.handlers); handlersLen != 0 {
		t.Fatalf("expected num of handlers to be 0; got %d", handlersLen)
	}

	if nodesLen := len(router.nodes); nodesLen != 0 {
		t.Fatalf("expected num of nodes to be 0; got %d", nodesLen)
	}
}

func TestAddRoute(t *testing.T) {
	r := newRouter()

	t.Run("existence route", func(st *testing.T) {
		emptyHandler := func(c *Context) {}

		tt := []struct {
			method string
			path   string
			key    string
		}{
			{http.MethodGet, "/", "GET-/"},
			{http.MethodGet, "/about", "GET-/about"},
			{http.MethodGet, "/downloads/*", "GET-/downloads/*"},
			{http.MethodPost, "/articles/:id", "POST-/articles/:id"},
		}

		for _, tc := range tt {
			r.addRoute(tc.method, tc.path, emptyHandler)

			if _, ok := r.handlers[tc.key]; !ok {
				st.Errorf("expected key %s to be exists in handlers", tc.key)
			}
		}
	})

	t.Run("handler count", func(st *testing.T) {
		firstHandler := func(c *Context) {}
		secondHandler := func(c *Context) {}
		r.addRoute(http.MethodGet, "/", firstHandler, secondHandler)

		route, ok := r.handlers["GET-/"]

		if !ok {
			st.Fatalf("expected route GET-/ to found; got not found")
		}

		if handlerCount := len(route); handlerCount != 2 {
			st.Errorf("expected handler count to be 2; got %d", handlerCount)
		}
	})

}

func TestFindRoute(t *testing.T) {
	r := newRouter()

	emptyHandler := func(c *Context) {}

	tt := []struct {
		name         string
		method       string
		urlPattern   string
		requestedURL string
		params       map[string]string
	}{
		{"root url", http.MethodGet, "/", "/", map[string]string{}},
		{"one parameter", http.MethodGet, "/users/:id", "users/1", map[string]string{"id": "1"}},
		{"one parameter with static path on last url", http.MethodGet, "/users/:id/about", "users/1/about", map[string]string{"id": "1"}},
		{"two parameter", http.MethodGet, "/users/:id/about/:section", "users/1/about/jobs", map[string]string{"id": "1", "section": "jobs"}},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(st *testing.T) {
			r.addRoute(tc.method, tc.urlPattern, emptyHandler)
			node, params := r.findRoute(tc.method, tc.requestedURL)
			if node == nil {
				st.Errorf("expected route to be found; got not found")
			}

			if node.urlPattern != tc.urlPattern {
				st.Errorf("expected found url to be %s; got %s", tc.urlPattern, node.urlPattern)
			}

			if paramsLen := len(params); paramsLen != len(tc.params) {
				st.Errorf("expected params length to be %d; got %d", len(tc.params), paramsLen)
			}

			for key, param := range params {
				if param != tc.params[key] {
					st.Errorf("expected param %s to be %s; got %s", key, tc.params[key], param)
				}
			}
		})
	}
}

func TestDefaultRouteHandler(t *testing.T) {
	r := newRouter()

	if r.defaultHandler != nil {
		t.Fatalf("expected default handler to be nil; got %T", r.defaultHandler)
	}

	tt := []struct {
		name             string
		method           string
		url              string
		responseCode     int
		responseText     string
		useCustomDefault bool // if it's true, this will modify default route handler using defaultCode & defaultText.
		defaultCode      int
		defaultText      string
	}{
		{"nano default route handler", http.MethodGet, "/", http.StatusNotFound, "nano/1.0 not found", false, 0, ""},
		{"set custom default route handler", http.MethodGet, "/", http.StatusOK, "it's works", true, http.StatusOK, "it's works"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(st *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				log.Fatalf("could not create http request: %v", err)
			}

			if tc.useCustomDefault {
				r.defaultHandler = func(c *Context) {
					c.String(tc.defaultCode, tc.defaultText)
				}
			}

			ctx := newContext(rec, req)
			r.serveDefaultHandler(ctx)

			if code := rec.Code; code != tc.responseCode {
				st.Fatalf("expected response code to be %d; got %d", tc.responseCode, code)
			}

			if body := rec.Body.String(); body != tc.responseText {
				st.Errorf("expected default handle response to be %s got %s", tc.responseText, body)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	r := newRouter()
	r.addRoute(http.MethodGet, "/hello/:name", func(c *Context) {
		c.String(http.StatusOK, "hello %s", c.Param("name"))
	})
	r.addRoute(http.MethodGet, "/d/*path", func(c *Context) {
		c.String(http.StatusOK, "downloading %s", c.Param("path"))
	})

	tt := []struct {
		name         string
		method       string
		url          string
		responseCode int
		responseText string
	}{
		{"not found handler", http.MethodGet, "/unregistered/path", http.StatusNotFound, "nano/1.0 not found"},
		{"not found on exist path but wrong method", http.MethodPost, "/hello/foo", http.StatusNotFound, "nano/1.0 not found"},
		{"echo parameter", http.MethodGet, "/hello/foo", http.StatusOK, "hello foo"},
		{"echo asterisk wildcard parameter", http.MethodGet, "/d/static/app.js", http.StatusOK, "downloading static/app.js"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(st *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				log.Fatalf("could not create http request: %v", err)
			}

			rec := httptest.NewRecorder()
			ctx := newContext(rec, req)
			r.handle(ctx)

			if code := rec.Code; code != tc.responseCode {
				st.Fatalf("expected response code to be %d; got %d", tc.responseCode, code)
			}

			if body := rec.Body.String(); body != tc.responseText {
				st.Errorf("expected %s as response text; got %v", tc.responseText, body)
			}
		})
	}
}
