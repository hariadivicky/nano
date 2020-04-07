<p align="center"><img width="320px" src="https://raw.githubusercontent.com/hariadivicky/logo/master/nano-logo-color.png"></p>

# Nano HTTP Multiplexer

[![Go Report Card](https://goreportcard.com/badge/github.com/hariadivicky/nano)](https://goreportcard.com/report/github.com/hariadivicky/nano) ![GitHub issues](https://img.shields.io/github/issues/hariadivicky/nano) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/hariadivicky/nano) ![GitHub](https://img.shields.io/github/license/hariadivicky/nano)

Nano is a simple & elegant HTTP multiplexer written in Go (Golang). It features REST API with Go net/http performance. If you need a minimalist, productivity and already familiar with net/http package, Nano is great choice.

## Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Usages](#api-usages)
  - [Using GET,POST,PUT, and DELETE](#using-get-post-put-and-delete)
    - [Router Parameter](#router-parameter)
    - [Grouping Routes](#grouping-routes)
    - [Writing Middleware](#writing-middleware)
    - [Using Middleware](#using-middleware)
    - [Middleware Group](#middleware-group)
    - [Nano Context](#nano-context)
      - [Request](#request)
      - [Response](#response)
- [Users](#users)

## Installation

To install Nano package, you need to install Go and set your Go workspace first.

- The first need [Go](https://golang.org/) installed (**version 1.12+ is required**), then you can use the below Go command to install Nano.

```sh
go get -u github.com/hariadivicky/nano
```

- Import it in your code:

```go
import "github.com/hariadivicky/nano"
```

- Import `net/http`. This is required.

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

```sh
# run example.go and visit http://localhost:8080 on your browser
$ go run example.go
```

## API Usages

You can find a number of ready-to-run examples at [Nano examples directory](https://github.com/hariadivicky/nano-examples).

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

### Grouping Routes

You could use grouping routes that have same prefix or using same middlewares

```go
app := nano.New()

// simple endpoint to print hello world.
app.GET("/", func(c *nano.Context) {
 c.String(http.StatusOK, "hello world\n")
})

// path: /api/v1
apiv1 := app.Group("/api/v1")
{
 // path: /api/v1/finances
 finance := apiv1.Group("/finances")
 {
  // path: /api/v1/finances/report/1
  finance.GET("/report/:period", reportByPeriodHandler)
  // path: /api/v1/finances/expenses/voucher
  finance.POST("/expenses/voucher", createExpenseVoucherHandler)
 }

 // path: /api/v1/users
 users := apiv1.Group("/users")
 {
  // path: /api/v1/users
  users.POST("/", registerNewUserHandler)
  // path: /api/v1/users/1/detail
  users.GET("/:id/detail", showUserDetailHandler)
 }
}
```

### Writing Middleware

Middleware implement nano.HandlerFunc, you could forward request by calling `c.Next()`

```go
// LoggerMiddleware functions to log every request.
func LoggerMiddleware() nano.HandlerFunc {
 return func(c *nano.Context) {
  // before middleware.
  start := time.Now()
  log.Println(c.Method, c.Path)

  // forward to next handler.
  c.Next()

  // after middleware.
  log.Printf("request complete in %s", time.Since(start))
 }
}
```

### Using Middleware

Using middleware on certain route

```go
app.GET("/change-password", verifyTokenMiddleware(), changePasswordHandler)
```

You could chaining middleware or handler

```go
app.GET("/secret", verifyStepOne(), verifyStepTwo(), grantAccessHandler, logChangeHandler)
```

### Middleware Group

Using middleware on router group

```go
app := nano.New()
// apply to all routes.
app.Use(globalMiddleware())

v1 := app.Group("/v1")
v1.Use(onlyForV1())
{
 // will apply to v1 routes.
}
```

### Nano Context

Nano Context is wrapper for http request and response. this example will use c variable as type of `*nano.Context`

#### Request

Get request `method` & `path`

```go
log.Println(c.Method, c.Path)
```

Get field value from request body

```go
username := c.FormValue("username")
```

Get url query

```go
page := c.Query("page")
```

#### Response

Set response header & content type

```go
c.SetHeader("X-Powered-By", "Nano HTTP Multiplexer")
c.SetContetType("image/png")
```

Plain text response

```go
c.String(http.StatusNotFound, "404 not found: %s", c.Path)
```

JSON response (with nano object wrapper)

```go
c.JSON(http.StatusOK, nano.H{
 "status":  "pending",
 "message": "data stored",
})
```

HMTL response

```go
c.HTML(http.StatusOK, "<h1>Hello There!</h1>")
```

Binary response

```go
c.Data(http.StatusOK, binaryData)
```

## Users

Awesome project lists using [Nano](https://github.com/hariadivicky/nano) web framework.

- [Coming soon](https://github.com/hariadivicky/coming-soon): A local retail shop written in Go.
