package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Step 1: Load configuration
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Step 2: Register telemetry (metrics)
	if config.Telemetry.Enabled {
		RegisterTelemetry()
	}

	// Step 3: Initialize router
	router := InitializeRoutes(config)

	// Step 4: Add `/metrics` endpoint if telemetry is enabled
	if config.Telemetry.Enabled && config.Telemetry.MetricsEndpoint != "" {
		router.Handle(config.Telemetry.MetricsEndpoint, promhttp.Handler())
		log.Printf("Prometheus metrics available at %s", config.Telemetry.MetricsEndpoint)
	}

	// Step 5: Define the HTTP server
	srv := &http.Server{
		Addr:         config.Server.Address,
		Handler:      router, // Use the updated router
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Step 6: Start the server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", config.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Step 7: Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Step 8: Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
