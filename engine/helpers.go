package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
)

// writeJSON writes a JSON response to the client.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// getEnv fetches an environment variable or returns a fallback value if not set.
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func generateFreshID() string {
	n := 5
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}

func (calc *calculator) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	calc.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (calc *calculator) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// decodeCT decodes a base64-String und extracts it, if it is a gzip-Blob ( 0x1f 0x8b).
func decodeCT(b64 string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	// gzip-Magic-Header
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		gr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return io.ReadAll(gr)
	}

	return raw, nil
}

func getBody(r *http.Request) ([]byte, error) {
	var rc io.ReadCloser = r.Body
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		rc = gz
	}
	return io.ReadAll(rc)
}

// Helper function to convert []byte results to Base64-encoded strings
func encodeResultsToBase64(results map[string][]byte) map[string]string {
	encodedResults := make(map[string]string)
	for key, value := range results {
		encodedResults[key] = base64.StdEncoding.EncodeToString(value)
	}
	return encodedResults
}
