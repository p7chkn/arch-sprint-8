package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

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

func NewKeyCloakMiddleware(jwks *keyfunc.JWKS, allowedRoles ...string) func(next http.Handler) http.Handler {
	roleMap := make(map[string]struct{}, len(allowedRoles))
	for _, v := range allowedRoles {
		roleMap[v] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Bearer token missing", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, jwks.Keyfunc)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Token is not valid", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Failed to parse claims", http.StatusUnauthorized)
				return
			}

			roles, ok := claims["realm_access"].(map[string]interface{})["roles"].([]interface{})
			if !ok {
				http.Error(w, "Roles not found in token", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), "roles", roles)

			for _, v := range roles {
				roleValue, ok := v.(string)
				if !ok {
					continue
				}
				if _, ok := roleMap[roleValue]; ok {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			http.Error(w, "Not allowed", http.StatusForbidden)
		})
	}
}
