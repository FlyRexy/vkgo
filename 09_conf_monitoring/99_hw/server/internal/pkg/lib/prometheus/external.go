package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ExternalRequestHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_requests_hits",
		},
		[]string{"path", "status", "service"},
	)
	ExternalRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "external_request_duration_seconds",
		},
		[]string{"path", "service"},
	)
)

func init() {
	prometheus.MustRegister(ExternalRequestHits, ExternalRequestDuration)
}
