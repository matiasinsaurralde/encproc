package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go" // Import for JWT package (use the appropriate import path)
)

// Secret key for signing the JWT token (should be kept secure, e.g., as an env variable)
var secretKey = []byte("m9Lk5RgBq23rTpqZn8A1F9Us4qaMphzd1knmn1H3p6A=")

//For testing:m9Lk5RgBq23rTpqZn8A1F9Us4qaMphzd1knmn1H3p6A=

func (calc *calculator) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		calc.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

// JWT Middleware to check Authorization header
func (calc *calculator) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")

		// If the header is missing or does not contain "Bearer <token>"
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return
		}

		// Extract the token part from the Authorization header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		token, err := validateToken(tokenString)
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// If valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// Validate the JWT token
func validateToken(tokenString string) (*jwt.Token, error) {
	// Parse the token using the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC SHA256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	return token, err
}
