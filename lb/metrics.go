package lb

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "load_balancer_requests_total",
			Help: "Total number of requests handled by the load balancer",
		},
		[]string{"method", "path"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "load_balancer_request_duration_seconds",
			Help:    "Duration of requests handled by the load balancer",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}
