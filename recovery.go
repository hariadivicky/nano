package nano

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

// Recovery middleware functions to recover when panic was fired.
func Recovery() HandlerFunc {
	return func(c *Context) {

		// defered call
		defer func() {
			if recovered := recover(); recovered != nil {
				err, ok := recovered.(error)

				if !ok {
					err = fmt.Errorf("%v", recovered)
				}

				// Create 1kb stack size.
				stacks := make([]byte, 1024)
				length := runtime.Stack(stacks, true)

				// print error and stack trace.
				log.Printf("[recovered] %v\n\nTrace %s\n", err, stacks[:length])

				// response
				c.String(http.StatusInternalServerError, "500 Internal Server Error")
			}
		}()

		c.Next()
	}
}
