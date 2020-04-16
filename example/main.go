package main

import (
	"net/http"

	"github.com/hariadivicky/nano"
)

type person struct {
	Name string
	Age  int
}

func main() {
	app := nano.New()

	// simple endpoint to print hello world.
	app.POST("/person", func(c *nano.Context) {
		if !c.IsJSON() {
			c.String(http.StatusBadRequest, "server only accept json request.")
		}

		form := new(person)
		err := c.ParseJSONBody(form)

		if err != nil {
			c.String(http.StatusBadRequest, "bad request.")
		}

		c.String(http.StatusOK, "hello %s\n your age is %d \n", form.Name, form.Age)
	})

	app.Run(":8080")
}
