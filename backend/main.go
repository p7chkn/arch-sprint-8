package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func getReports(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("HTTP Caracola"))
}

func main() {
	const serverAddr string = ":8000"
	done := make(chan struct{})

	router := http.NewServeMux()

	router.Handle("/reports", keycloakAuthMiddleware(http.HandlerFunc(getReports)))

	corsRouter := corsMiddleware(router)

	server := http.Server{
		Addr:    serverAddr,
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
