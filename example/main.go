package main

import (
	"net/http"

	"github.com/hariadivicky/nano"
)

func main() {
	app := nano.New()

	// simple endpoint to print hello world.
	app.GET("/", func(c *nano.Context) {
		c.String(http.StatusOK, "hello world\n")
	})

	// return product id.
	app.GET("/products/:id", func(c *nano.Context) {
		c.JSON(http.StatusOK, nano.H{
			"product_id": c.Param("id"),
		})
	})

	app.Run(":8080")
}
