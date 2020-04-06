package nano

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Context defines nano request - response context.
type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Method  string
	Path    string
	Params  map[string]string
}

// newContext is Context constructor.
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: r,
		Writer:  w,
		Method:  r.Method,
		Path:    r.URL.Path,
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

// Query functions to get url query.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
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
