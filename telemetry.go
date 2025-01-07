package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Define metrics globally
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	// Other metrics can be added here...
)

// RegisterTelemetry registers all Prometheus metrics
func RegisterTelemetry() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
}
