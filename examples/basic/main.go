// Basic example with middleware and parameterized route.
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
