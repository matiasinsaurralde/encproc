package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/collapsinghierarchy/encproc/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
)

type calculator struct {
	logger        *slog.Logger
	calc_model    models.EncProcModelAPI
	jWTMiddleware *jWTMiddleware
	agg_map       sync.Map // concurrent map: key string -> *aggregator
	aux_map       sync.Map // concurrent map: key string -> auxData
	thumbsUpCount int64
}

//	@title			Encproc API engine
//	@version		0.1
//	@description	Encrypted Processing API engine

//	@contact.name	Encproc Dev Team
//	@contact.url	https://pseudocrypt.site
//	@contact.email	encproc@gmail.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		pseudocrypt.site
// @BasePath	/
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	//------------------ Init the CORS Middleware -----------------------
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	//------------------ Init the database connections -----------------------
	addr := getEnv("API_ADDR", ":8080")
	metricsAddr := getEnv("METRICS_ADDR", ":9000")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "mydb")
	dbUser := getEnv("DB_USER", "myuser")
	dbPassword := getEnv("DB_PASSWORD", "mypassword")
	jwt_sk := getEnv("SECRET_KEY", "m9Lk5RgBq23rTpqZn8A1F9Us4qaMphzd1knmn1H3p6A=")

	// Read optional TLS material from the environment.
	certFile := os.Getenv("TLS_CERT_FILE") // e.g. "/certs/fullchain.pem"
	keyFile := os.Getenv("TLS_KEY_FILE")   // e.g. "/certs/privkey.pem"

	// Build the connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := openDB(dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	calc := &calculator{
		logger:     logger,
		calc_model: &models.EncProcModel{DB: db}, // This struct must implement EncProcModelAPI
	}

	/*
		* TLS Configuration
			(05/2025) Regarding PQC: https://github.com/golang/go/issues/64537 --> Therefore we will have to stick with what's available now.
	*/
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}

	calc.calc_model.InitializeTables()
	apiMux := calc.routes()

	jwtMW := &jWTMiddleware{secretKey: []byte(jwt_sk)}
	calc.jWTMiddleware = jwtMW

	apiSrv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         addr,
		Handler:      c.Handler(apiMux),
		TLSConfig:    tlsConfig,
	}

	// ───────────────────────────── middleware / routing ──────────────────────────
	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics",
		promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, // publish promhttp_* stats here
			promhttp.Handler(),           // exporter
		),
	)

	// ───────────────────────────── metrics server ──────────────────────────────
	metricsSrv := &http.Server{Addr: metricsAddr, Handler: metricsMux}

	// ─────────────────────────── context + errgroup ─────────────────────
	// 1) cancel ctx on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2) all goroutines derive from that ctx
	g, ctx := errgroup.WithContext(ctx)
	// ──────────────────────────── start the servers ────────────────────────────
	// API listener
	g.Go(func() error {
		// shut down gracefully on Ctrl-C / docker stop
		go func() { <-ctx.Done(); _ = apiSrv.Shutdown(context.Background()) }()
		if certFile != "" && keyFile != "" {
			calc.logger.Info("TLS enabled on " + addr)
			return apiSrv.ListenAndServeTLS(certFile, keyFile)
		}
		calc.logger.Info("Plain HTTP on " + addr)
		return apiSrv.ListenAndServe()
	})

	// metrics listener
	g.Go(func() error {
		go func() { <-ctx.Done(); _ = metricsSrv.Shutdown(context.Background()) }()
		return metricsSrv.ListenAndServe()
	})

	//calc.logger.Info("API on :443, metrics on :9000 — press Ctrl-C to stop")
	calc.logger.Info("API on " + addr + ", metrics on " + metricsAddr + " — press Ctrl-C to stop")
	// wait for error OR signal
	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		calc.logger.Error("server error", "error", err)
	}

}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
