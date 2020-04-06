# Nano HTTP Multiplexer

<img align="right" width="169px" src="https://raw.githubusercontent.com/hariadivicky/logo/master/nano-logo-color.png">

Nano is a simple & elegant HTTP multiplexer written in Go (Golang). It features REST API with Go net/http performance. If you need a minimalist, productivity and already familiar with net/http package, Nano is great choice.

## Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [API Usages](#api-usages)
    - [Using GET,POST,PUT,PATCH,DELETE and OPTIONS](#using-get-post-put-patch-delete-and-options)
    - [Router parameter](#router-parameter)
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
	app := nano.New()

	// simple endpoint to print hello world.
	app.GET("/", func(c *nano.Context) {
		c.String(http.StatusOK, "hello world\n")
	})

	app.Run(":8080")
}
```

```
# run example.go and visit http://localhost:8080 on your browser
$ go run example.go
```

## API Usages

You can find a number of ready-to-run examples at [Nano examples repository](https://github.com/hariadivicky/nano-examples).

### Using GET, POST, PUT, and DELETE

```go
func main() {
	// Creates a nano router
	app := nano.New()

	app.GET("/someGet", getHandler)
	app.POST("/somePost", postHandler)
	app.PUT("/somePut", putHandler)
	app.DELETE("/someDelete", deleteHandler)

    // Run apps.
	app.Run(":8080")
}
```

### Router parameter

Get route parameter

```go
func main() {
	app := nano.New()

	// This handler will match /products/1 but will not match /products/ or /products
	app.GET("/products/:productId", func(c *nano.Context) {
        productId := c.Param("productId") //string
        c.String(http.StatusOK, "you requested %s", productId)
	})
}
```

## Users

Awesome project lists using [Nano](https://github.com/hariadivicky/nano) web framework.

* [Coming soon](https://github.com/hariadivicky/coming-soon): A local retail shop written in Go.
