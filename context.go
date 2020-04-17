package nano

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Context defines nano request - response context.
type Context struct {
	Request  *http.Request
	Writer   http.ResponseWriter
	Method   string
	Path     string
	Origin   string
	Params   map[string]string
	handlers []HandlerFunc
	cursor   int // used for handlers stack.
}

// newContext is Context constructor.
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: r,
		Writer:  w,
		Method:  r.Method,
		Path:    r.URL.Path,
		Origin:  r.Header.Get("Origin"),
		cursor:  -1,
	}
}

// Next functions to move cursor to the next handler stack.
func (c *Context) Next() {
	// moving cursor.
	c.cursor++

	if c.cursor < len(c.handlers) {
		c.handlers[c.cursor](c)
	}
}

// Status functions to set http status code response.
func (c *Context) Status(statusCode int) {
	c.Writer.WriteHeader(statusCode)
}

// SetHeader functions to set http response header.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// SetContentType functions to set http content type response header.
func (c *Context) SetContentType(contentType string) {
	c.SetHeader("Content-Type", contentType)
}

// Param functions to get request parameter.
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// PostForm functions to form body field.
func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

// PostFormDefault return default value when form body field is empty.
func (c *Context) PostFormDefault(key string, defaultValue interface{}) interface{} {
	v := c.PostForm(key)

	if v == "" {
		return defaultValue
	}

	return v
}

// Query functions to get url query.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault return default value when url query is empty
func (c *Context) QueryDefault(key string, defaultValue interface{}) interface{} {
	v := c.Query(key)

	if v == "" {
		return defaultValue
	}

	return v
}

// IsJSON returns true when client send json body.
func (c *Context) IsJSON() bool {
	contentType := c.Request.Header.Get("Content-Type")
	return contentType == "application/json"
}

// ParseJSONBody functions to parse json request body.
func (c *Context) ParseJSONBody(body interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(&body)
}

// ExpectJSON returns true when client request json response
func (c *Context) ExpectJSON() bool {
	acceptHeader := c.Request.Header.Get("Accept")
	return strings.Contains(acceptHeader, "application/json")
}

// JSON functions to write json response.
func (c *Context) JSON(statusCode int, object interface{}) {
	c.SetContentType("application/json")
	c.Status(statusCode)

	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(object); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// String functions to write plain text response.
func (c *Context) String(statusCode int, template string, value ...interface{}) {
	c.SetContentType("text/plain")
	c.Status(statusCode)

	text := fmt.Sprintf(template, value...)

	c.Writer.Write([]byte(text))
}

// HTML functions to write html response.
func (c *Context) HTML(statusCode int, html string) {
	c.SetContentType("text/html")
	c.Status(statusCode)
	c.Writer.Write([]byte(html))
}

// Data functions to write binary response.
func (c *Context) Data(statusCode int, binary []byte) {
	c.Status(statusCode)
	c.Writer.Write(binary)
}
