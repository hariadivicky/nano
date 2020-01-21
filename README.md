# Nano HTTP Router

<img align="right" width="169px" src="https://raw.githubusercontent.com/hariadivicky/logo/master/nano-logo-color.png">

Nano is a simple HTTP router written in Go (Golang). It features REST API with Go net/http performance. If you need a minimalist, productivity and already familiar with net/http package, Nano is great choice.

## Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [API Usages](#api-usages)
    - [Using GET,POST,PUT,PATCH,DELETE and OPTIONS](#using-get-post-put-patch-delete-and-options)
    - [Router parameter](#router-parameter)
    - [Grouping](#grouping)
    - [Writing middleware](#writing-middleware)
    - [Using middleware](#using-middleware)
- [Users](#users)

## Installation

To install Nano package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed (**version 1.12+ is required**), then you can use the below Go command to install Nano.

```sh
$ go get -u github.com/hariadivicky/nano
```

2. Import it in your code:

```go
import "github.com/hariadivicky/nano"
```

3. Import `net/http`. This is required.

```go
import "net/http"
```

## Quick start
 
```sh
# assume the following codes in example.go file
$ cat example.go
```

```go
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
```

```
# run example.go and visit 0.0.0.0:8080 (for windows "localhost:8080") on browser
$ go run example.go
```

## API Usages

You can find a number of ready-to-run examples at [Nano examples repository](https://github.com/hariadivicky/nano-examples).

### Using GET, POST, PUT, PATCH, DELETE and OPTIONS

```go
func main() {
	// Creates a nano router
	router := nano.New()

	router.Get("/someGet", getting)
	router.Post("/somePost", posting)
	router.Put("/somePut", putting)
	router.Delete("/someDelete", deleting)

    // Attach router as mux
	http.ListenAndServe(":8080", router)
}
```

### Router parameter

Get route parameter

```go
func main() {
	router := nano.New()

	// This handler will match /products/1 but will not match /products/ or /products
	router.Get("/products/:productId", func(w http.ResponseWriter, r *http.Request) {
        params := router.ParseParameter(r)
        
        // return string
        productId := params.Get("productId")
        fmt.Fprint(w, productId)
	})
}
```

Set/replace route parameter

```go
func main() {
	router := nano.New()

	router.Get("/increment/:num", func(w http.ResponseWriter, r *http.Request) {
        params := router.ParseParameter(r)
        num, _ := strconv.Atoi(params.Get("num"))

        // replace num value.
        params.Set("num", num+1)

        fmt.Fprint(w, num)
	})
}
```

### Grouping

Prefix group routes

```go
func main() {
    // base URL is /api
    router := nano.New("/api")
    
    router.Get("/", welcome)

    vendor := router.Group("/vendors")
    {
        vendor.Get("/", vendorList)
        // this endpoint can be access at /api/vendors/1
        vendor.Get("/:id", vendorDetail)
        vendor.Put("/:id", vendorUpdate)
    }
}
```

### Writing middleware

A middleware receive and return type of `http.HandlerFunc`

```go
func accessLogger(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // before middleware
        log.Println(r.Method, r.URL.String())

        // forward to next handler
        next(w, r)

        // after middleware
        log.Println("finish")
    }
}
```

### Using middleware

You can use middleware at router group level or even at a single route

```go
func main() {
    vendor := router.Group("/vendors")
    // group level middleware
    vendor.Use(vendorAccessLogger, authenticatedOnly)
    {
        // single route level middleware
        vendor.Get("/", vendorList).Use(xmlHeader)
        vendor.Get("/:id", vendorDetail)
        vendor.Put("/:id", vendorUpdate)
    }
}
```

## Users

Awesome project lists using [Nano](https://github.com/hariadivicky/nano) web framework.

* [Coming soon](https://github.com/hariadivicky/coming-soon): A local retail shop written in Go.
