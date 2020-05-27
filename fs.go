package nano

import (
	"net/http"
)

// fileServerHandler to handle static file server.
func fileServerHandler(routerPrefix, baseURL string, rootDir http.FileSystem) HandlerFunc {
	return func(c *Context) {
		prefix := baseURL + "/"
		// if current file server not in root group, append router group prefix to baseurl.
		if routerPrefix != "" {
			prefix = routerPrefix + baseURL + "/"
		}

		fs := http.FileServer(rootDir)
		// remove static prefix of url.
		fileServer := http.StripPrefix(prefix, fs)

		// we will check existence of file,
		// if current requested file doesn't exists, we will send not found as response.
		file, err := rootDir.Open(c.Param("filepath"))
		if err != nil {
			c.String(http.StatusNotFound, "file not found")
			return
		}

		stat, err := file.Stat()
		if err != nil {
			panic(err)
		}
		file.Close()

		// disable directory listing.
		if stat.IsDir() {
			c.String(http.StatusForbidden, "access forbidden")
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
