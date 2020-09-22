<p align="center"><img width="320px" src="https://raw.githubusercontent.com/hariadivicky/logo/master/nano-logo-color.png"></p>

# Nano HTTP Multiplexer

[![Go Report Card](https://goreportcard.com/badge/github.com/hariadivicky/nano)](https://goreportcard.com/report/github.com/hariadivicky/nano) ![GitHub issues](https://img.shields.io/github/issues/hariadivicky/nano) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/hariadivicky/nano) ![GitHub](https://img.shields.io/github/license/hariadivicky/nano)

Nano is a simple & elegant HTTP multiplexer written in Go (Golang). It features REST API with Go net/http performance. If you need a minimalist, productivity, and love simplicity, Nano is great choice.

## Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Usages](#api-usages)
  - [Using HEAD, OPTIONS, GET, POST, PUT, PATCH, and DELETE](#using-head-options-get-post-put-patch-and-delete)
  - [Default Route Handler](#default-route-handler)
  - [Route Parameter](#route-parameter)
  - [Static File Server](#static-file-server)
  - [Request Binding](#request-binding)
    - [Bind URL Query](#bind-url-query)
    - [Bind Multipart Form](#bind-multipart-form)
    - [Bind JSON](#bind-json)
    - [Error Binding](#error-binding)
  - [Grouping Routes](#grouping-routes)
  - [Writing Middleware](#writing-middleware)
  - [Using Middleware](#using-middleware)
  - [Middleware Group](#middleware-group)
  - [Nano Context](#nano-context)
    - [Request](#request)
    - [Response](#response)
- [Nano Middlewares](#nano-middlewares)
  - [Recovery Middleware](#recovery-middleware)
  - [CORS Middleware](#cors-middleware)
  - [Gzip Middleware](#gzip-middleware)
- [Users](#users)
- [License](#license)

## Installation

To install Nano package, you need to install Go and set your Go workspace first.

- The first need [Go](https://golang.org/) installed (**version 1.11+ is required**), then you can use the below Go command to install Nano.

```sh
go get -u github.com/hariadivicky/nano
```

- Import it in your code:

```go
import "github.com/hariadivicky/nano"
```

- Import `net/http`. This is optional for http status code like `http.StatusOK`.

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

You can find a ready-to-run examples at [Todo List Example](https://github.com/hariadivicky/nano-example).

### Using HEAD, OPTIONS, GET, POST, PUT, PATCH, and DELETE

```go
func main() {
    // Creates a nano router
    app := nano.New()

    app.HEAD("/someHead", headHandler)
    app.OPTIONS("/someOptions", optionsHandler)
    app.GET("/someGet", getHandler)
    app.POST("/somePost", postHandler)
    app.PUT("/somePut", putHandler)
    app.PATCH("/somePatch", patchHandler)
    app.DELETE("/someDelete", deleteHandler)

    // Run apps.
    app.Run(":8080")
}
```

### Default Route Handler

You could register you own default handler, the default handler will called when there is not matching route found. If you doesn't set the default handler, nano will register default 404 response as default handler.

```go
app.Default(func(c *nano.Context) {
    c.JSON(http.StatusNotFound, nano.H{
        "status": "error",
        "message": "Oh no, this API endpoint does not exists."
    })
})
```

### Route Parameter

Get route parameter using `c.Param(key)`

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

### Static File Server

You could use `*nano.Static()` function to serve static files like html, css, and js in your server.

```go
func main() {
    app := nano.New()

    assetDir := http.Dir("./statics")
    app.Static("/assets", assetDir)

    // now your static files are accessible via http://yourhost.com/assets/*
}
```

### Request Binding

To use request binding you must provide `form` tag to each field in your struct. You can also add the validation rules using `validate` tag. to see more about available `validate` tag value, visit [Go Validator](https://github.com/go-playground/validator/)

```go
type Address struct {
    Street     string `form:"street" json:"street"`
    PostalCode string `form:"postal_code" json:"postal_code"`
    CityID     int    `form:"city_id" json:"city_id" validate:"required"`
}
```

By calling `Bind` function, it's will returns `*nano.BindingError` when an error occured due to deserialization error or validation error. The description about error fields will be stored in `err.Fields`.

```go

app.GET("/address", func(c *nano.Context) {
    var address Address
    if err := c.Bind(&address); err != nil {
        c.String(err.HTTPStatusCode, err.Message)
        return
    }

    c.String(http.StatusOK, "city id: %d, postal code: %s", address.CityID, address.PostalCode)
})

```

The `Bind` function automatically choose deserialization source based on your request Content-Type and request method. `GET` and `HEAD` methods will try to bind url query or urlencoded form. Otherwise, it will try to bind multipart form or json.

but if you want to manually choose the binding source, you could uses this functions below:

#### Bind URL Query

If you want to bind url query like `page=1&limit=50` or urlencoded form you could use `BindSimpleForm`

```go
var paging Pagination
err := c.BindSimpleForm(&paging)
```

#### Bind Multipart Form

You could use `BindMultipartForm` to bind request body with `multipart/form-data` type

```go
var post BlogPost
err := c.BindMultipartForm(&post)
```

#### Bind JSON

if you have request with `application/json` type, you could bind it using `BindJSON` function

```go
var schema ComplexSchema
err := c.BindJSON(&schema)
```

`BindJSON` can also parsing your `RFC3339` date/time format to another format by adding `time_format` in your field tag. You can read more at [jsontime](https://github.com/liamylian/jsontime) docs.

#### Error Binding

Each you call `Bind`, `BindSimpleForm`, `BindMultipartForm`, and `BindJSON` it's always returns `*nano.ErrorBinding,` except when binding success without any errors it returns `nil`. ErrorBinding has two field that are HTTPStatusCode & Message. Here is the details:

|   | HTTPStatusCode | Reason                                                          |
|---|----------------|-----------------------------------------------------------------|
| 1 | 500            | Conversion Error or Give non-pointer to target struct parameter |
| 2 | 420            | Validation Error                                                |
| 3 | 400            | Deserialization Error                                           |

`ErrorBinding.HTTPStatusCode` is useful to determine response code

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

Nano Context is wrapper for http request and response. this example will use `c` variable as type of `*nano.Context`

#### Request

Get request `method` & `path`

```go
log.Println(c.Method, c.Path)
```

Get field value from request body

```go
username := c.PostForm("username")
```

Get field value with default value from request body

```go
status := c.PostFormDefault("status", "active")
```

Get url query

```go
page := c.Query("page")
```

Get url query with default value

```go
page := c.QueryDefault("page", "1")
```

You could check if client need JSON response

```go
func main() {
    app := nano.New()

    // simple endpoint to print hello world.
    app.GET("/", func(c *nano.Context) {

        // client request json response.
        if c.ExpectJSON() {
            c.JSON(http.StatusOK, nano.H{
                "message": "hello world",
            })

            return
        }

        c.String(http.StatusOK, "hello world\n")
    })

    app.Run(":8080")
}
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

## Nano Middlewares

Nano has shipped with some default middleware like cors and recovery middleware.

### Recovery Middleware

Recovery middleware is functions to recover server when panic was fired.

```go
func main() {
    app := nano.New()
    app.Use(nano.Recovery())

    app.GET("/", func(c *nano.Context) {
        stack := make([]string, 0)

        c.String(http.StatusOK, "100th stack is %s", stack[99])
    })

    app.Run(":8080")
}
```

### CORS Middleware

This middleware is used to deal with cross-origin request.

```go
func main() {
    app := nano.New()

    // Only allow from :3000 and google.
    cors := nano.CORSWithConfig(nano.CORSConfig{
        AllowedOrigins: []string{"http://localhost:3000", "https://wwww.google.com"},
        AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
        AllowedHeaders: []string{nano.HeaderContentType, nano.HeaderAccept},
    })

    app.Use(cors)

    // ...
}
```

### Gzip Middleware

Gzip middleware is used for http response compression.

```go
func main() {
    app := nano.New()

    app.Use(nano.Gzip(gzip.DefaultCompression))

    // ...
}
```

don't forget to import `compress/gzip` package for compression level at this example. available compression levels are: `gzip.NoCompression`, `gzip.BestSpeed`, `gzip.BestCompression`, `gzip.DefaultCompression`, and `gzip.HuffmanOnly`

## Users

Awesome project lists using [Nano](https://github.com/hariadivicky/nano) web framework.

- [Coming soon](https://github.com/hariadivicky/coming-soon): A local retail shop written in Go.

## License

Nano using MIT license. Please read LICENSE files for more information about nano license.
