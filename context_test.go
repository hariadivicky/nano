package nano

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateNewContext(t *testing.T) {
	path := "/hello"
	method := http.MethodGet
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}

	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	if ctx.Path != path {
		t.Errorf("expected path to be %s; got %s", path, ctx.Path)
	}

	if ctx.Method != method {
		t.Errorf("expected method to be %s; got %s", method, ctx.Method)
	}

	if ctx.cursor != -1 {
		t.Errorf("expected cursor to be -1; got %d", ctx.cursor)
	}
}

func TestNext(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	emptyHandler := func(c *Context) {
		c.Next()
	}
	helloHandler := func(c *Context) {
		c.String(http.StatusOK, "ok")
	}

	r := newRouter()
	r.addRoute(http.MethodGet, "/", emptyHandler, emptyHandler, emptyHandler, helloHandler, emptyHandler)
	r.handle(ctx)

	if ctx.cursor != 3 {
		t.Fatalf("expected stack cursor to be 3; got %d", ctx.cursor)
	}

	if hlen := len(ctx.handlers); hlen != 5 {
		t.Errorf("expected total handler to be 5; got %d", hlen)
	}
}

func TestStatusCode(t *testing.T) {
	r := newRouter()
	r.addRoute(http.MethodGet, "/", func(c *Context) {
		c.Status(http.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	r.handle(ctx)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status code to be %d; got %d", http.StatusNoContent, rec.Code)
	}
}

func TestSetHeader(t *testing.T) {
	headers := map[string]string{
		"X-Powered-By": "nano/1.1",
		"X-Foo":        "Bar,Baz",
	}

	r := newRouter()
	r.addRoute(http.MethodGet, "/", func(c *Context) {
		for key, val := range headers {
			c.SetHeader(key, val)
		}

		c.String(http.StatusOK, "ok")
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	r.handle(ctx)

	for key, val := range headers {
		if head := rec.Header().Get(key); head != val {
			t.Errorf("expected header %s to be %s; got %s", key, val, head)
		}
	}
}

// GetRequestHeader returns header value by given key.
func TestGetRequestHeader(t *testing.T) {
	reqHeaders := [2]struct {
		Key   string
		Value string
	}{
		{HeaderContentType, MimeJSON},
		{HeaderAccept, MimeJSON},
	}

	r := newRouter()
	r.addRoute(http.MethodGet, "/", func(c *Context) {
		params := make([]string, 0)
		for _, header := range reqHeaders {
			val := c.GetRequestHeader(header.Key)
			params = append(params, fmt.Sprintf("%s:%s", header.Key, val))
		}
		c.String(http.StatusOK, strings.Join(params, ","))
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	for _, header := range reqHeaders {
		req.Header.Set(header.Key, header.Value)
	}

	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	r.handle(ctx)

	params := make([]string, 0)
	for _, rs := range reqHeaders {
		params = append(params, fmt.Sprintf("%s:%s", rs.Key, rs.Value))
	}

	expectedRes := strings.Join(params, ",")
	if body := rec.Body.String(); body != expectedRes {
		t.Fatalf("expected response body %s; got %s", expectedRes, body)
	}
}

func TestGetRouteParameter(t *testing.T) {
	tt := []struct {
		name        string
		method      string
		urlPattern  string
		requestURL  string
		paramNames  []string
		paramValues []string
	}{
		{"one parameter", http.MethodGet, "/hello/:name", "/hello/foo", []string{"name"}, []string{"foo"}},
		{"two parameters", http.MethodGet, "/hello/:name/show/:section", "/hello/bar/show/media", []string{"name", "section"}, []string{"bar", "media"}},
		{"one parameter with asterisk wildcard", http.MethodGet, "/d/:timeout/u/*path", "/d/30/u/files/nano.zip", []string{"timeout", "path"}, []string{"30", "files/nano.zip"}},
		// please consider the test ordering, because it's has added to same router instance.
		{"asterisk wildcard", http.MethodGet, "/d/u/*path", "/d/u/files/nano.zip", []string{"path"}, []string{"files/nano.zip"}},
	}

	r := newRouter()

	for _, tc := range tt {
		r.addRoute(tc.method, tc.urlPattern, func(c *Context) {
			params := make([]string, 0)
			for _, name := range tc.paramNames {
				params = append(params, c.Param(name))
			}

			c.String(http.StatusOK, strings.Join(params, ","))
		})

		req, err := http.NewRequest(tc.method, tc.requestURL, nil)
		if err != nil {
			log.Fatalf("could not make http request: %v", err)
		}

		rec := httptest.NewRecorder()
		ctx := newContext(rec, req)
		r.handle(ctx)

		expectedRes := strings.Join(tc.paramValues, ",")
		if body := rec.Body.String(); body != expectedRes {
			t.Errorf("expected parameter to be %s; got %s", expectedRes, body)
		}
	}
}

func TestResponse(t *testing.T) {
	r := newRouter()
	jsonHandler := func(c *Context) {
		c.JSON(http.StatusOK, H{
			"message": "ok",
		})
	}
	stringHandler := func(c *Context) {
		c.String(http.StatusOK, "ok")
	}
	htmlHandler := func(c *Context) {
		c.HTML(http.StatusOK, "<h1>ok</h1>")
	}
	jsonErrorHandler := func(c *Context) {
		c.JSON(http.StatusOK, make(chan int))
	}
	binaryHandler := func(c *Context) {
		c.Data(http.StatusOK, []byte("ok"))
	}

	tt := []struct {
		name        string
		url         string
		status      int
		handler     HandlerFunc
		contentType string
	}{
		{"string", "/string", http.StatusOK, stringHandler, MimePlainText},
		{"json", "/json", http.StatusOK, jsonHandler, MimeJSON},
		{"json with error", "/json/error", http.StatusInternalServerError, jsonErrorHandler, MimePlainText},
		{"html", "/html", http.StatusOK, htmlHandler, MimeHTML},
		{"data", "/data", http.StatusOK, binaryHandler, ""},
	}

	for _, tc := range tt {
		r.addRoute(http.MethodGet, tc.url, tc.handler)

		req, err := http.NewRequest(http.MethodGet, tc.url, nil)
		if err != nil {
			log.Fatalf("could not make http request: %v", err)
		}

		rec := httptest.NewRecorder()
		ctx := newContext(rec, req)
		r.handle(ctx)

		if rec.Code != tc.status {
			t.Fatalf("expected status code to be %d; got %d", tc.status, rec.Code)
		}

		if contentType := rec.Header().Get(HeaderContentType); contentType != tc.contentType {
			t.Errorf("expected content type to be %s; got %s", tc.contentType, contentType)
		}
	}
}

func TestIsJSON(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	req.Header.Add(HeaderContentType, MimeJSON)
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	if !ctx.IsJSON() {
		t.Errorf("expected IsJSON to be true; got %v", ctx.IsJSON())
	}
}

func TestExpectJSON(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	req.Header.Add(HeaderAccept, MimeJSON)
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	if !ctx.ExpectJSON() {
		t.Errorf("expected ExpectJSON to be true; got %v", ctx.ExpectJSON())
	}
}

func TestQuery(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/?name=foo", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	if name := ctx.Query("name"); name != "foo" {
		t.Errorf("expected query name to be foo; got %s", name)
	}

	if name := ctx.QueryDefault("name", "must ignored"); name != "foo" {
		t.Errorf("expected name to be foo; got %s", name)
	}

	if grade := ctx.QueryDefault("grade", "1"); grade != "1" {
		t.Errorf("expected query grade to be 1; got %s", grade)
	}
}

func TestForm(t *testing.T) {
	form := []byte("name=foo")
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(form))
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	req.Header.Add(HeaderContentType, MimeFormURLEncoded)
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	if name := ctx.PostForm("name"); name != "foo" {
		t.Errorf("expected name post field to be foo; got %s", name)
	}
	if name := ctx.PostFormDefault("name", "must be ignored"); name != "foo" {
		t.Errorf("expected name post field to be foo; got %s", name)
	}
	if grade := ctx.PostFormDefault("grade", "1"); grade != "1" {
		t.Errorf("expected grade post field to be 1; got %s", grade)
	}
}

func TestBind(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/?name=foo&gender=male", nil)
	if err != nil {
		log.Fatalf("could not make http request: %v", err)
	}
	rec := httptest.NewRecorder()
	ctx := newContext(rec, req)

	var person struct {
		Name   string `form:"name"`
		Gender string `form:"gender"`
	}

	// because we doesn't specify the request content type & we use GET method,
	// this should bind the url query.
	ctx.Bind(&person)

	if person.Name != "foo" {
		t.Errorf("expected person name to be foo; got %s", person.Name)
	}

	if person.Gender != "male" {
		t.Errorf("expected person gender to be male; got %s", person.Gender)
	}
}
