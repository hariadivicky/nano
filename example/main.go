package main

import (
	"fmt"
	"net/http"

	"github.com/hariadivicky/nano"
)

func main() {
	router := nano.New()

	// simple endpoint to print hello world.
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world!\n")
	})

	http.ListenAndServe(":8080", router)
}
