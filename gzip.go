package nano

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

// Gzip compression for http response.
// this compression works when client accept gzip in their request.
func Gzip(compressionLevel int) HandlerFunc {
	return func(c *Context) {
		// make sure if client request has gzip in accept-encoding header.
		if !strings.Contains(c.GetRequestHeader(HeaderAcceptEncoding), "gzip") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, compressionLevel)
		// this error may caused incorrect compression level value.
		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		c.SetHeader(HeaderContentEncoding, "gzip")
		defer gz.Close()

		gzWriter := &gzipWriter{c.Writer, gz}

		// replace default writter with Gzip Writer.
		c.Writer = gzWriter
		c.Next()
	}
}

// Write overrides default http response writer with gzip writter.
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// WriteHeader overrides response writer to delete content length.
// reference: https://github.com/labstack/echo/issues/444
// If Content-Length header is set, gzip probably writes the wrong number of bytes.
// We should delete the Content-Length header prior to writing the headers on a gzipped response.
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del(HeaderContentLength)
	g.ResponseWriter.WriteHeader(code)
}
