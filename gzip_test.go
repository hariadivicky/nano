package nano

import (
	"compress/gzip"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	app := New()
	app.Use(Gzip(gzip.DefaultCompression))

	app.GET("/", func(c *Context) {
		c.String(http.StatusOK, "hello world")
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}

	req.Header.Add(HeaderAcceptEncoding, "gzip")

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if encoding := rec.Header().Get(HeaderContentEncoding); encoding != "gzip" {
		t.Errorf("expected encoding to be gzip; got %s", encoding)
	}
}

func TestGzipWithoutAcceptEncoding(t *testing.T) {
	app := New()
	app.Use(Gzip(gzip.DefaultCompression))

	app.GET("/", func(c *Context) {
		c.String(http.StatusOK, "hello world")
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if encoding := rec.Header().Get(HeaderContentEncoding); encoding == "gzip" {
		t.Errorf("expected encoding not to be gzip; got %s", encoding)
	}
}

func TestGzipWithWrongCompressionLevel(t *testing.T) {
	app := New()

	app.Use(Gzip(10))

	app.GET("/", func(c *Context) {
		c.String(http.StatusOK, "hello world")
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not create http request: %v", err)
	}
	req.Header.Add(HeaderAcceptEncoding, "gzip")

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected response code to be 500; got %v", rec.Code)
	}

	if encoding := rec.Header().Get(HeaderContentEncoding); encoding == "gzip" {
		t.Errorf("expected encoding not to be gzip; got %s", encoding)
	}
}
