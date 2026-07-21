package external

import (
	"net/http"
	"server/internal/pkg/lib/prometheus"
	"strconv"
	"time"
)

func NewClient(service string) *http.Client {
	return &http.Client{
		Transport: &observedTransport{
			next:    http.DefaultTransport,
			service: service,
		},
	}
}

type observedTransport struct {
	next    http.RoundTripper
	service string
}

func (ot *observedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	res, err := ot.next.RoundTrip(req)

	status := "0"
	if res != nil {
		status = strconv.Itoa(res.StatusCode)
	}

	prometheus.ExternalRequestHits.WithLabelValues(req.URL.Path, status, ot.service).Inc()
	prometheus.ExternalRequestDuration.WithLabelValues(req.URL.Path, ot.service, status).Observe(time.Since(start).Seconds())

	return res, err
}
