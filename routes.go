package main

import (
	"net/http"

	"github.com/justinas/alice"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (calc *calculator) routes() http.Handler {

	mux := http.NewServeMux()

	//------------------ Register the Static Files -----------------------
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/static/swagger/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	mux.Handle("GET /{$}", fileServer)
	mux.HandleFunc("POST /create-stream", calc.authMiddleware(calc.createStream))
	mux.HandleFunc("POST /contribute/aggregate/{id}", calc.contributeAggregate)
	mux.HandleFunc("GET /snapshot/aggregate/{id}", calc.returnAggregate)
	mux.HandleFunc("GET /public-key/{id}", calc.getPublicKey)

	// Create a middleware chain
	standard := alice.New(calc.logRequest)

	// Return the 'standard' middleware chain followed by the servemux.
	return standard.Then(mux)
}
