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
		Origin:  r.Header.Get(HeaderOrigin),
		cursor:  -1,
	}
}

// Next is functions to move cursor to the next handler stack.
func (c *Context) Next() {
	// moving cursor.
	c.cursor++

	if c.cursor < len(c.handlers) {
		c.handlers[c.cursor](c)
	}
}

// Status is functions to set http status code response.
func (c *Context) Status(statusCode int) {
	c.Writer.WriteHeader(statusCode)
}

// SetHeader is functions to set http response header.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// GetRequestHeader returns header value by given key.
func (c *Context) GetRequestHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetContentType is functions to set http content type response header.
func (c *Context) SetContentType(contentType string) {
	c.SetHeader(HeaderContentType, contentType)
}

// Param functions is to get request parameter.
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// PostForm is functions to form body field.
func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

// PostFormDefault return default value when form body field is empty.
func (c *Context) PostFormDefault(key string, defaultValue string) string {
	v := c.PostForm(key)

	if v == "" {
		return defaultValue
	}

	return v
}

// Query is functions to get url query.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault return default value when url query is empty
func (c *Context) QueryDefault(key string, defaultValue string) string {
	v := c.Query(key)

	if v == "" {
		return defaultValue
	}

	return v
}

// Bind request body into defined user struct.
// This function help you to automatic binding based on request Content-Type & request method.
// If you want to chooose binding method manualy, you could use :
// BindSimpleForm to bind urlencoded form & url query,
// BindMultipartForm to bind multipart/form data,
// and BindJSON to bind application/json request body.
func (c *Context) Bind(targetStruct interface{}) *BindingError {
	return bind(c, targetStruct)
}

// IsJSON returns true when client send json body.
func (c *Context) IsJSON() bool {
	return c.GetRequestHeader(HeaderContentType) == MimeJSON
}

// ParseJSONBody is functions to parse json request body.
func (c *Context) ParseJSONBody(body interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(&body)
}

// ExpectJSON returns true when client request json response
func (c *Context) ExpectJSON() bool {
	return strings.Contains(c.GetRequestHeader(HeaderAccept), MimeJSON)
}

// JSON is functions to write json response.
func (c *Context) JSON(statusCode int, object interface{}) {
	c.SetContentType(MimeJSON)
	c.Status(statusCode)

	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(object); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// String is functions to write plain text response.
func (c *Context) String(statusCode int, template string, value ...interface{}) {
	c.SetContentType(MimePlainText)
	c.Status(statusCode)

	text := fmt.Sprintf(template, value...)

	c.Writer.Write([]byte(text))
}

// HTML is functions to write html response.
func (c *Context) HTML(statusCode int, html string) {
	c.SetContentType(MimeHTML)
	c.Status(statusCode)
	c.Writer.Write([]byte(html))
}

// Data is functions to write binary response.
func (c *Context) Data(statusCode int, binary []byte) {
	c.Status(statusCode)
	c.Writer.Write(binary)
}
