package middlewares

import (
	"context"
	"fmt"
	"log"
	"mnctech-restapi/cmd/rest-api/auth"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5" // Adjust based on your JWT library
)

// JWTMiddleware checks for a valid JWT token and extracts the user ID.
func JWTMiddleware(accessTokenKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the JWT token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}
			// Split the token string to get the token part
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse the token
			token, err := jwt.ParseWithClaims(tokenString, &auth.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				return accessTokenKey, nil
			})

			if err != nil {
				log.Println("Error parsing token:", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Check if the token is valid and extract claims
			if claims, ok := token.Claims.(*auth.CustomClaims); ok && token.Valid {
				// Store UID in context
				ctx := context.WithValue(r.Context(), auth.UserIDKey, claims.UID)
				r = r.WithContext(ctx) // Update the request with the new context

				fmt.Printf("Claims: %+v\n", claims.UID) // Print claims for debugging
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			// If the token is valid, call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
