package nano

import (
	"fmt"
	"net/http"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// Bag .
type Bag struct {
	data map[string]interface{}
}

// NewBag creates new bag instance.
func NewBag() *Bag {
	return &Bag{
		data: make(map[string]interface{}),
	}
}

// Set data to bag.
func (b *Bag) Set(key string, data interface{}) {
	b.data[key] = data
}

// Get data by given key.
func (b *Bag) Get(key string) interface{} {
	if data, ok := b.data[key]; ok {
		return data
	}

	return nil
}

// Context defines nano request - response context.
type Context struct {
	Request    *http.Request
	Writer     http.ResponseWriter
	Method     string
	Path       string
	Origin     string
	Params     map[string]string
	handlers   []HandlerFunc
	Bag        *Bag
	cursor     int // used for handlers stack.
	validator  *validator.Validate
	translator ut.Translator
}

// newContext is Context constructor.
func newContext(w http.ResponseWriter, r *http.Request) *Context {

	trans := newTranslator()
	validator := newValidator(trans)

	return &Context{
		Request:    r,
		Writer:     w,
		Method:     r.Method,
		Path:       r.URL.Path,
		Origin:     r.Header.Get(HeaderOrigin),
		cursor:     -1,
		Bag:        NewBag(),
		validator:  validator,
		translator: trans,
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

// IsJSON returns true when client send json body.
func (c *Context) IsJSON() bool {
	return c.GetRequestHeader(HeaderContentType) == MimeJSON
}

// ExpectJSON returns true when client request json response,
// since this function use string.Contains, value ordering in Accept values doesn't matter.
func (c *Context) ExpectJSON() bool {
	return strings.Contains(c.GetRequestHeader(HeaderAccept), MimeJSON)
}

// JSON is functions to write json response.
func (c *Context) JSON(statusCode int, object interface{}) {
	rs, err := json.Marshal(object)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.SetContentType(MimeJSON)
	c.Status(statusCode)
	c.Writer.Write(rs)
}

// String is functions to write plain text response.
func (c *Context) String(statusCode int, template string, value ...interface{}) {
	c.SetContentType(MimePlainText)
	c.Status(statusCode)

	text := fmt.Sprintf(template, value...)

	c.Writer.Write([]byte(text))
}

// File will returns static file as response.
func (c *Context) File(statusCode int, filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
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
