package main

import (
	"net/http"

	"github.com/hariadivicky/nano"
)

func main() {
	app := nano.New()

	// simple endpoint to print hello world.
	app.POST("/", func(c *nano.Context) {
		c.String(http.StatusOK, "hello world\n")
	})

	app.Run(":8080")
}
