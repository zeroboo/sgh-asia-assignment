package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zeroboo.webservice/handler"
)

func newRouter() http.Handler {
	mux := http.NewServeMux()

	//Use appropriate method
	mux.HandleFunc("POST /save", handler.Save)
	return mux
}

const (
	serverAddr      = ":8080"
	readTimeout     = 10 * time.Second
	writeTimeout    = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      newRouter(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("starting server on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Wait for system signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// gracefully shut down
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}
