// Command sgh-asia-ai starts the payment service and demonstrates the worker pool.
package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
	"zeroboo.payment/handler"
	"zeroboo.payment/service"
)

func main() {

	// Init logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Init database
	dbStr := os.Getenv("MYSQL_CONNECTION_STRING") // e.g. "user:password@tcp(127.0.0.1:3306)/payment?parseTime=true"
	if dbStr == "" {
		dbStr = "root:@tcp(127.0.0.1:3306)/payment?parseTime=true"
	}
	db, err := sql.Open("mysql", dbStr)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	logger.Info("Connected to SQL", "MYSQL_CONNECTION_STRING", dbStr)
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Init gin router
	paymentService := service.NewPaymentService(db)
	r := gin.New()
	h := handler.NewPaymentHandler(paymentService, logger)
	h.RegisterRoutes(r)

	// Setup server
	addr := ":8080"
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Payment service starting on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Listen to system signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("server is stopping...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown http server failed: error=%w", err)
	}
	log.Println("server stopped!")
}
