// Command orderengine starts an in-memory order matching engine behind a
// small REST API. See README.md for what this is, how to run it, and how to
// exercise it with curl.
package main

import (
	"log"
	"net/http"

	"orderengine/internal/api"
	"orderengine/internal/engine"
)

func main() {
	book := engine.NewOrderBook()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, book)

	log.Println("order matching engine listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
