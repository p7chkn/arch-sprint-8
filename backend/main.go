package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/MicahParks/keyfunc"
	"golang.org/x/exp/rand"
)

type Report struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	Author    string `json:"author"`
	Status    string `json:"status"`
}

func getReports(w http.ResponseWriter, r *http.Request) {
	report := Report{
		ID:        rand.Intn(1000),
		Title:     "Employee Feedback",
		Timestamp: "2024-12-06T14:35:19Z",
		Author:    "Alice",
		Status:    "complete",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, "Failed to encode report", http.StatusInternalServerError)
	}
}

func main() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	jwks, err := keyfunc.Get(cfg.KeyCloakCertURL, keyfunc.Options{
		RefreshInterval: 1 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create JWKS from URL %s: %v", cfg.KeyCloakCertURL, err)
	}

	serverAddress := fmt.Sprintf(":%s", cfg.HostPort)

	done := make(chan struct{})

	router := http.NewServeMux()

	keycloakAuthMiddleware := NewKeyCloakMiddleware(jwks, "prothetic_user")

	router.Handle("/reports", keycloakAuthMiddleware(http.HandlerFunc(getReports)))

	corsRouter := corsMiddleware(router)

	server := http.Server{
		Addr:    serverAddress,
		Handler: corsRouter,
	}
	server.RegisterOnShutdown(func() {
		defer func() {
			done <- struct{}{}
		}()
	})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		_, _ = fmt.Fprintf(os.Stderr, "HTTP server error %v\n", err)
		close(done)
	}
	<-done
}
