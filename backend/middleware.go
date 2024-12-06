package main

import (
	"context"
	"fmt"
	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

var jwksURL = "http://keycloak:8080/realms/reports-realm/protocol/openid-connect/certs"

var jwks *keyfunc.JWKS

func init() {
	var err error
	jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: 1 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create JWKS from URL %s: %v", jwksURL, err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func keycloakAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Extract the token from the header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Bearer token missing", http.StatusUnauthorized)
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Token is not valid", http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Failed to parse claims", http.StatusUnauthorized)
			return
		}

		// Extract roles from claims
		roles, ok := claims["realm_access"].(map[string]interface{})["roles"].([]interface{})
		if !ok {
			http.Error(w, "Roles not found in token", http.StatusForbidden)
			return
		}

		if !slices.Contains(roles, "prothetic_user") {
			http.Error(w, "Not allowed", http.StatusForbidden)
		}

		// Add roles to the request context for further use
		ctx := context.WithValue(r.Context(), "roles", roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
