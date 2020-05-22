package nano

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUseMiddleware(t *testing.T) {
	app := New()

	emptyHandler := func(c *Context) {}

	app.Use(emptyHandler, emptyHandler, emptyHandler)

	if mlen := len(app.middlewares); mlen != 3 {
		t.Errorf("expect num of middlewares to be 3; got %d", mlen)
	}
}

func TestGroup(t *testing.T) {
	app := New()

	api := app.Group("/api")
	if api.prefix != "/api" {
		t.Errorf("expected group prefix to be /api; got %s", api.prefix)
	}

	finance := api.Group("/finance")
	if finance.prefix != "/api/finance" {
		t.Errorf("expected group prefix to be /api/finance; got %s", api.prefix)
	}
}

func TestRouteRegistration(t *testing.T) {
	app := New()

	emptyHandler := func(c *Context) {}
	app.GET("/", emptyHandler)
	app.POST("/", emptyHandler)
	app.PUT("/", emptyHandler)
	app.DELETE("/", emptyHandler)

	if hlen := len(app.router.handlers); hlen != 4 {
		t.Errorf("expected num of registered routes to be 4; got %d", hlen)
	}
}

func TestDefaultHandler(t *testing.T) {
	app := New()

	if app.router.defaultHandler != nil {
		t.Fatalf("expected initial value of default handler to be nil")
	}

	t.Run("set default handler", func(st *testing.T) {
		app.Default(func(c *Context) {
			c.String(http.StatusOK, "ok")
		})

		if app.router.defaultHandler == nil {
			st.Errorf("expected default handler to be setted; got %v", app.router.defaultHandler)
		}
	})

	t.Run("set default handler when it already set", func(st *testing.T) {
		err := app.Default(func(c *Context) {
			c.String(http.StatusOK, "ok")
		})

		if err != ErrDefaultHandler {
			st.Errorf("expected result to be ErrDefaultHandler; got %v", err)
		}
	})
}

func TestServeHTTP(t *testing.T) {
	app := New()
	app.GET("/", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected response code to be 200; got %d", rec.Code)
	}

	if body := rec.Body.String(); body != "ok" {
		t.Errorf("expected response text to be ok; got %s", body)
	}
}
