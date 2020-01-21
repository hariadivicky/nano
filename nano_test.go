package nano_test

import (
	"fmt"
	"net/http"

	"github.com/hariadivicky/nano"
)

// Basic usages to create hello world.
func Example() {
	router := nano.New()

	// simple endpoint to print hello world.
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world!\n")
	})

	http.ListenAndServe(":8080", router)
}
