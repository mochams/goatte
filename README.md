# goatte 🚀

`goatte` is a lightweight HTTP library for building web services in Go.  
It's built on top of Go’s standard library and is inspired by [net/http](https://pkg.go.dev/net/http), [django](https://github.com/django/django), and [chi](https://github.com/go-chi/chi).  
Thanks for checking it out.

No extra dependencies — just Go.  
100% compatible with net/http.

[![Go Reference](https://pkg.go.dev/badge/github.com/mochams/goatte.svg)](https://pkg.go.dev/github.com/mochams/goatte)
[![License](https://img.shields.io/github/license/mochams/goatte)](https://github.com/mochams/goatte?tab=BSD-3-Clause-1-ov-file)

## Features

- Easy to use and learn

- Clean and simple routing

- Built-in support for middleware

- URL reversing (name your routes and look them up)

- Serve static files easily

- Organize routes with sub-routers

- 100% compatible with net/http

- Clean and minimal API

- Zero dependencies (only uses the Go standard library)

## Installation

```bash
go get github.com/mochams/goatte@latest
```

## Quick start

Create a web server with middleware, and a parameterized route:

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/mochams/goatte"
)

// Sample middleware that logs messages before and after the request is handled
var ExampleMiddleware = func(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Before handler (middleware)")
        next.ServeHTTP(w, r)
        fmt.Println("After handler (middleware)")
    })
}

func main() {
    middleware := []goatte.Middleware{ExampleMiddleware}
    router := goatte.NewRouter("main", middleware...)

    // Also fine
    // router := goatte.NewRouter("main", ExampleMiddleware)

    router.Get("/users/{name}/", "user-detail", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("..Handling request in handler...")
        w.Write([]byte("Hello, " + r.PathValue("name")))
    })

    log.Println("Server listening on http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", router))
}
```

Start the server

```bash
go run .
```

In a new shell, call the endpoint:

```bash
curl -i http://localhost:3000/users/goatte/
```

You should see this in the terminal running your server:

```bash
Before handler (middleware)
..Handling request in handler...
After handler (middleware)
```

And this in the response:

```bash
HTTP/1.1 200 OK
...

Hello, goatte%   
```

## Examples

Looking for more examples?  
Check out the [examples/](https://github.com/mochams/goatte/tree/develop/examples) folder for more examples

## Contributing

Contributions are welcome!  
If you have ideas, bug fixes, or improvements, feel free to open a pull request.

## Acknowledgments

`goatte` is inspired by the simplicity of Go’s standard net/http, Chi’s routing, and the best ideas of Django framework.
