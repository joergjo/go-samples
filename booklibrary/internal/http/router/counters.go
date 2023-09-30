package router

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

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "booklibrary_request_duration_seconds",
			Help:    "A histogram of latencies for booklibrary API requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)

	// responseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "booklibrary_response_size_bytes",
			Help:    "A histogram of response sizes for booklibrary API requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)
)

func instrument(handlerFunc http.HandlerFunc, handlerName string) http.Handler {
	return promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": handlerName}),
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize, handlerFunc),
			),
		),
	)
}
