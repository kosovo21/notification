package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts HTTP requests by method, path, and status code.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration observes request latency in seconds.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// MessagesPublishedTotal counts messages published to the queue.
	MessagesPublishedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "messages_published_total",
			Help: "Total number of messages published to the queue.",
		},
	)

	// MessagesProcessedTotal counts messages processed by the worker.
	MessagesProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "messages_processed_total",
			Help: "Total number of messages processed by the worker.",
		},
		[]string{"platform", "result"},
	)
)
