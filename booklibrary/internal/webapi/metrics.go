package webapi

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	inFlightGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "booklibrary_in_flight_requests",
		Help: "A gauge of requests currently being served by the booklibrary API.",
	})

	counter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "booklibrary_api_requests_total",
			Help: "A counter for requests to the the booklibrary API.",
		},
		[]string{"code", "method"},
	)

	duration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "booklibrary_request_duration_seconds",
			Help:    "A histogram of latencies for booklibrary API requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)

	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "booklibrary_response_size_bytes",
			Help:    "A histogram of response sizes for booklibrary API requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)
)

func metricsFor(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerInFlight(inFlightGauge,
			promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": name}),
				promhttp.InstrumentHandlerCounter(counter,
					promhttp.InstrumentHandlerResponseSize(responseSize, next),
				),
			),
		)
	}
}
