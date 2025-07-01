package main

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var (
	httpReqs = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "All HTTP-Requests sorted by method, path and code",
		},
		[]string{"method", "path", "code"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Response time until header write",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "code"},
	)
)

// instrument wraps an http.Handler to instrument it with Prometheus metrics.
// It uses the provided path to label the metrics, which allows for more granular tracking.
func instrument(path string, h http.Handler) http.Handler {
	c := httpReqs.MustCurryWith(prometheus.Labels{"path": path})
	d := httpDuration.MustCurryWith(prometheus.Labels{"path": path})

	return promhttp.InstrumentHandlerDuration(
		d,
		promhttp.InstrumentHandlerCounter(c, h),
	)
}

func (calc *calculator) routes() http.Handler {
	mux := http.NewServeMux()

	//-------------------------------------------------- Static
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("GET /static/", instrument("/static/", http.StripPrefix("/static", fileServer)))

	//-------------------------------------------------- Swagger
	mux.Handle("GET /docs/", instrument(
		"/docs/",
		httpSwagger.Handler(
			httpSwagger.URL("/static/swagger.json"),
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("none"),
			httpSwagger.DomID("swagger-ui"),
		),
	))

	//-------------------------------------------------- REST-API
	mux.Handle("GET /{$}", instrument("/", fileServer))
	mux.Handle("POST /create-stream", instrument("/create-stream", http.HandlerFunc(calc.createStream)))
	mux.Handle("POST /contribute/aggregate", instrument("/contribute/aggregate", http.HandlerFunc(calc.contributeAggregate)))
	mux.Handle("GET /snapshot/aggregate/{id}", instrument("/snapshot/aggregate/{id}", http.HandlerFunc(calc.returnAggregate)))
	mux.Handle("GET /public-key/{id}", instrument("/public-key/{id}", http.HandlerFunc(calc.getPublicKey)))
	mux.Handle("GET /stream/{id}", instrument("/stream/{id}", http.HandlerFunc(calc.streamDetails)))
	mux.Handle("GET /stream/{id}/{display}", instrument("/stream/{id}/{display}", http.HandlerFunc(calc.streamDetails)))
	mux.Handle("GET /thumbs-up", instrument("/thumbs-up[GET]", http.HandlerFunc(calc.getThumbsUp)))
	mux.Handle("POST /thumbs-up", instrument("/thumbs-up[POST]", http.HandlerFunc(calc.incrementThumbsUp)))

	// Middleware chain
	standard := alice.New(calc.logRequest)
	return standard.Then(mux)
}
