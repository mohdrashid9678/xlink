package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Sub-millisecond to multi-second latency distribution buckets optimized for URL shorteners
var defaultLatencyBuckets = []float64{
	0.0001, // 100 microseconds (L1 memory hit)
	0.0005, // 500 microseconds (L2 Redis hit)
	0.001,  // 1 millisecond
	0.0025, // 2.5 milliseconds
	0.005,  // 5 milliseconds
	0.01,   // 10 milliseconds
	0.025,  // 25 milliseconds
	0.05,   // 50 milliseconds
	0.1,    // 100 milliseconds
	0.25,   // 250 milliseconds
	0.5,    // 500 milliseconds
	1.0,    // 1 second
	2.5,    // 2.5 seconds
	5.0,    // 5 seconds
}

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight *prometheus.GaugeVec
	CacheOperations  *prometheus.CounterVec
}

var DefaultMetrics = NewMetrics(prometheus.DefaultRegisterer)

func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	return &Metrics{
		RequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "xlink",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests partitioned by method, route handler, and response status code.",
			},
			[]string{"method", "handler", "status_code"},
		),
		RequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "xlink",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency distributions partitioned by method, route handler, and status code.",
				Buckets:   defaultLatencyBuckets,
			},
			[]string{"method", "handler", "status_code"},
		),
		RequestsInFlight: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "xlink",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Current number of active HTTP requests being processed partitioned by handler.",
			},
			[]string{"handler"},
		),
		CacheOperations: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "xlink",
				Subsystem: "cache",
				Name:      "operations_total",
				Help:      "Total number of multi-tier cache lookups partitioned by tier (l1/l2) and result (hit/miss).",
			},
			[]string{"tier", "result"},
		),
	}
}
