package main

import (
	"net/http"

	"github.com/justinas/alice"
)

// The routes() method returns a servemux containing our application routes.
func (calc *calculator) routes() http.Handler {

	mux := http.NewServeMux()

	//------------------ Register the Static Files -----------------------
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.Handle("GET /{$}", fileServer)                                            // Serve index.html for root path
	mux.HandleFunc("POST /create-stream", calc.authMiddleware(calc.createStream)) //curl -X POST http://localhost:9000/create-stream -H "Authorization: Bearer <your-jwt-token>"
	mux.HandleFunc("POST /contribute/aggregate/{id}", calc.contributeAggregate)
	mux.HandleFunc("GET /snapshot/aggregate/{id}", calc.returnAggregate)
	mux.HandleFunc("GET /public-key/{id}", calc.getPublicKey)

	// Create a middleware chain
	standard := alice.New(calc.logRequest)

	// Return the 'standard' middleware chain followed by the servemux.
	return standard.Then(mux)
}
