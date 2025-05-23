package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/cors"
)

type calculator struct {
	logger        *slog.Logger
	calc_model    *EncProcModel
	jWTMiddleware *jWTMiddleware
	agg_map       sync.Map // concurrent map: key string -> *aggregator
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

	// Allow all origins, adjust as needed
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	//------------------ Init the database connections -----------------------
	addr := getEnv("ADDR", ":443")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "mydb")
	dbUser := getEnv("DB_USER", "myuser")
	dbPassword := getEnv("DB_PASSWORD", "mypassword")
	jwt_sk := getEnv("SECRET_KEY", "m9Lk5RgBq23rTpqZn8A1F9Us4qaMphzd1knmn1H3p6A=")

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
		calc_model: &EncProcModel{DB: db},
	}

	/*
		* TLS Configuration
			(05/2025) Regarding PQC-TLS: https://github.com/golang/go/issues/64537 --> Therefore we will have to stick with what's available now.
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

	calc.calc_model.initializeTables()
	mux := calc.routes()

	jwtMW := &jWTMiddleware{secretKey: []byte(jwt_sk)}
	calc.jWTMiddleware = jwtMW

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         addr,
		Handler:      c.Handler(mux),
		TLSConfig:    tlsConfig,
	}

	calc.logger.Info("starting server on :443")

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	logger.Error(err.Error())
	os.Exit(1)
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
