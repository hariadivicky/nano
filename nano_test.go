package nano_test

import (
	"net/http"

	"github.com/hariadivicky/nano"
)

// Basic usages to create hello world.
func Example() {
	app := nano.New()

	// simple endpoint to print hello world.
	app.GET("/", func(c *nano.Context) {
		c.String(http.StatusOK, "hello world")
	})

	app.Run(":8080")
}
