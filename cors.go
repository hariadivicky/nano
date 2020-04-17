package nano

// This cross-origin sharing standard is used to enable cross-site HTTP requests for:

// Invocations of the XMLHttpRequest or Fetch APIs in a cross-site manner, as discussed above.
// Web Fonts (for cross-domain font usage in @font-face within CSS), so that servers can deploy TrueType fonts that can only be cross-site loaded and used by web sites that are permitted to do so.
// WebGL textures.
// Images/video frames drawn to a canvas using drawImage().
import (
	"net/http"
	"strings"
)

// CORSConfig define nano cors middleware configuration.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// CORS struct.
type CORS struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// parseRequestHeader functions to split header string to array of headers.
func parseRequestHeader(header string) []string {

	// request does not provide field Access-Control-Request-Header.
	if header == "" {
		return []string{}
	}

	// only requested one header.
	if !strings.Contains(header, ",") {
		return []string{header}
	}

	result := make([]string, 0)

	for _, part := range strings.Split(header, ",") {
		result = append(result, strings.Trim(part, " "))
	}

	return result
}

// SetAllowedOrigins functions to fill/replace all allowed origins.
func (cors *CORS) SetAllowedOrigins(origins []string) {
	cors.allowedOrigins = origins
}

// SetAllowedMethods functions to fill/replace all allowed methods.
func (cors *CORS) SetAllowedMethods(methods []string) {
	cors.allowedMethods = methods
}

// SetAllowedHeaders functions to fill/replace all allowed headers.
func (cors *CORS) SetAllowedHeaders(headers []string) {
	cors.allowedHeaders = headers
}

// AddAllowedHeader functions to append method to allowed list.
func (cors *CORS) AddAllowedHeader(header string) {
	cors.allowedHeaders = append(cors.allowedHeaders, header)
}

// AddAllowedMethod functions to append method to allowed list.
func (cors *CORS) AddAllowedMethod(method string) {
	cors.allowedMethods = append(cors.allowedMethods, method)
}

// AddAllowedOrigin functions to append method to allowed list.
func (cors *CORS) AddAllowedOrigin(origin string) {
	cors.allowedOrigins = append(cors.allowedOrigins, origin)
}

// isAllowAllOrigin returns true when there is * wildcrad in the origin list.
func (cors *CORS) isAllowAllOrigin() bool {
	for _, origin := range cors.allowedOrigins {
		if origin == "*" {
			return true
		}
	}

	return false
}

// isOriginAllowed returns true when origin found in allowed origin list.
func (cors *CORS) isOriginAllowed(requestOrigin string) bool {
	for _, origin := range cors.allowedOrigins {
		if origin == requestOrigin || origin == "*" {
			return true
		}
	}

	return false
}

// isMethodAllowed returns true when method found in allowed method list.
func (cors *CORS) isMethodAllowed(requestMethod string) bool {
	for _, method := range cors.allowedMethods {
		if method == requestMethod {
			return true
		}
	}

	return false
}

// mergeMethods functions to stringify the allowed method list.
func (cors *CORS) mergeMethods() string {
	// when there is found * wildcard in the list, so just return it.
	for _, method := range cors.allowedMethods {
		if method == "*" {
			return method
		}
	}

	return strings.Join(cors.allowedMethods, ", ")
}

// isAllHeaderAllowed returns true when there is * wildcrad in the allowed header list.
func (cors *CORS) isAllHeaderAllowed() bool {
	for _, header := range cors.allowedHeaders {
		if header == "*" {
			return true
		}
	}

	return false
}

// areHeadersAllowed functions to check are all requested headers are allowed
func (cors *CORS) areHeadersAllowed(requestedHeaders []string) bool {
	// alway return true if there is no control header.
	if cors.isAllHeaderAllowed() {
		return true
	}

	for _, requestedHeader := range requestedHeaders {
		allowed := false

		for _, allowedHeader := range cors.allowedHeaders {
			if allowedHeader == requestedHeader {
				allowed = true
			}
		}

		if !allowed {
			return false
		}
	}

	return true
}

// handlePrefilghtRequest functions to handle cross-origin preflight request.
func (cors *CORS) handlePrefilghtRequest(c *Context) {
	if c.Origin == "" {
		return
	}

	if !cors.isOriginAllowed(c.Origin) {
		return
	}

	requestedMethod := c.Request.Header.Get("Access-Control-Request-Method")
	if !cors.isMethodAllowed(requestedMethod) {
		return
	}

	requestedHeader := c.Request.Header.Get("Access-Control-Request-Header")
	requestedHeaders := parseRequestHeader(requestedHeader)

	if len(requestedHeaders) > 0 {
		if !cors.areHeadersAllowed(requestedHeaders) {
			return
		}
	}

	// vary must be set.
	c.SetHeader("Vary", "Origin, Access-Control-Request-Methods, Access-Control-Request-Header")

	if cors.isAllowAllOrigin() {
		c.SetHeader("Access-Control-Allow-Origin", "*")
	} else {
		c.SetHeader("Access-Control-Allow-Origin", c.Origin)
	}

	c.SetHeader("Access-Control-Allow-Methods", cors.mergeMethods())

	if len(requestedHeader) > 0 {
		c.SetHeader("Access-Control-Allow-Header", requestedHeader)
	}
}

// handleSimpleRequest functions to handle simple cross origin request
// see:
func (cors *CORS) handleSimpleRequest(c *Context) {
	if c.Origin == "" {
		return
	}

	if !cors.isOriginAllowed(c.Origin) {
		return
	}

	// vary must be set.
	c.SetHeader("Vary", "Origin")

	if cors.isAllowAllOrigin() {
		c.SetHeader("Access-Control-Allow-Origin", "*")
	} else {
		c.SetHeader("Access-Control-Allow-Origin", c.Origin)
	}
}

// Handle corss-origin request
// The Cross-Origin Resource Sharing standard works by adding new HTTP headers that allow servers
// to describe the set of origins that are permitted to read that information using a web browser.
// Additionally, for HTTP request methods that can cause side-effects on server's data
// (in particular, for HTTP methods other than GET, or for POST usage with certain MIME types),
// the specification mandates that browsers "preflight" the request,
// soliciting supported methods from the server with an HTTP OPTIONS request method,
// and then, upon "approval" from the server, sending the actual request with the actual HTTP request method.
// Servers can also notify clients whether "credentials" (including Cookies and HTTP Authentication data) should be sent with requests.
func (cors *CORS) Handle(c *Context) {
	// preflighted requests first send an HTTP request by the OPTIONS method to the resource on the other domain,
	// in order to determine whether the actual request is safe to send.
	// Cross-site requests are preflighted like this since they may have implications to user data.
	if c.Method == http.MethodOptions && c.Request.Header.Get("Access-Control-Request-Method") != "" {
		cors.handlePrefilghtRequest(c)
		return
	}

	// Some requests don’t trigger a CORS preflight. Those are called “simple requests”,
	// though the Fetch spec (which defines CORS) doesn’t use that term.
	// A request that doesn’t trigger a CORS preflight—a so-called “simple request”
	cors.handleSimpleRequest(c)

	c.Next()
}

// CORSWithConfig returns cors middleware.
func CORSWithConfig(config CORSConfig) HandlerFunc {

	cors := new(CORS)

	// create default value for all configuration field.
	// default value is allowed for all origin, methods, and headers.
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodGet}
	}

	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}

	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"*"}
	}

	cors.SetAllowedMethods(config.AllowedMethods)
	cors.SetAllowedOrigins(config.AllowedOrigins)
	cors.SetAllowedHeaders(config.AllowedHeaders)

	return cors.Handle
}
