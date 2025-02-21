package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/rs/cors"
)

type calculator struct {
	logger        *slog.Logger
	calc_model    *EncProcModel
	jWTMiddleware *jWTMiddleware
	agg_map       sync.Map // concurrent map: key string -> *aggregator
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Allow all origins, adjust as needed
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Replace "*" with your specific domains if needed
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	//------------------ Init the database connections -----------------------
	// Fetch database configuration from environment variables
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
	calc.calc_model.initializeTables()
	mux := calc.routes()
	jwtMW := &jWTMiddleware{secretKey: []byte(jwt_sk)}
	calc.jWTMiddleware = jwtMW

	// Start the server
	calc.logger.Info("starting server on :8080")
	err = http.ListenAndServe(":8080", c.Handler(mux))
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
