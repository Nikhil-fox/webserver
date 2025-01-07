package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

func InitializeRoutes(config *Config) *mux.Router {
	router := mux.NewRouter()

	// Create rate limiter if enabled
	if config.RateLimiting.Enabled {
		log.Printf("Rate limiting enabled with max requests: %d and time window: %s", config.RateLimiting.MaxRequests, config.RateLimiting.TimeWindow)
		timeWindow, _ := time.ParseDuration(config.RateLimiting.TimeWindow)
		rateLimiter := newRateLimiter(config.RateLimiting.MaxRequests, timeWindow)
		router.Use(RateLimitingMiddleware(rateLimiter)) // Apply globally
	}

	// Health check endpoint
	router.HandleFunc(config.API.Routes["health"], HealthCheckHandler).Methods("GET")

	// API endpoints
	api := router.PathPrefix("/api/" + config.API.Version).Subrouter()
	api.HandleFunc(config.API.Routes["get_items"], GetItemsHandler).Methods("GET")
	api.HandleFunc(config.API.Routes["create_item"], CreateItemHandler).Methods("POST")

	return router
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		RequestCounter.WithLabelValues(r.Method, path, "200").Inc()
		RequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	itemsMutex sync.RWMutex
	items      = []Item{
		{ID: "1", Name: "Item One"},
		{ID: "2", Name: "Item Two"},
	}
)

func GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	itemsMutex.RLock()
	defer itemsMutex.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var newItem Item
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itemsMutex.Lock()
	items = append(items, newItem)
	itemsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newItem)
}
