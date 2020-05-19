package nano

import (
	"net/http"
	"strings"
)

// BindingError is an error wrapper.
// HTTPStatusCode will set to 422 when there is error on validation,
// 400 when client sent unsupported/without Content-Type header, and
// 500 when targetStruct is not pointer or type conversion is fail.
type BindingError struct {
	HTTPStatusCode int
	Message        string
}

// bind request body to defined user struct.
// This function help you to automatic binding based on request Content-Type & request method
func bind(c *Context, targetStruct interface{}) *BindingError {
	contentType := c.GetRequestHeader(HeaderContentType)

	// if client request using POST, PUT, & PATCH we will try to bind request using simple form (urlencoded & url query),
	// multipart form, and JSON. if you need both binding e.g. to bind multipart form & url query,
	// this method doesn't works. you should call BindSimpleForm & BindMultipartForm manually from your handler.
	if c.Method == "POST" || c.Method == "PUT" || c.Method == "PATCH" || contentType != "" {
		if strings.Contains(contentType, MimeFormURLEncoded) {
			return BindSimpleForm(c.Request, targetStruct)
		}

		if strings.Contains(contentType, MimeMultipartForm) {
			return BindMultipartForm(c.Request, targetStruct)
		}

		if c.IsJSON() {
			return BindJSON(c.Request, targetStruct)
		}

		return &BindingError{
			HTTPStatusCode: http.StatusBadRequest,
			Message:        "unknown content type of request body",
		}
	}

	// when client request using GET method, we will serve binding using simple form.
	// it's can binding url-encoded form & url query data.
	return BindSimpleForm(c.Request, targetStruct)
}
