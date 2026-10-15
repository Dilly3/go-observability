package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

func NewMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),

		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wallet_http_requests_total",
				Help: "Total completed HTTP requests, excluding metrics scrapes.",
			},
			[]string{"method", "route", "status"},
		),

		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "wallet_http_request_duration_seconds",
				Help: "Server-side HTTP request duration in seconds.",
				Buckets: []float64{
					0.001,
					0.005,
					0.01,
					0.025,
					0.05,
					0.1,
					0.25,
					0.5,
					1,
					2.5,
					5,
				},
			},
			[]string{"method", "route"},
		),

		inFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "wallet_http_requests_in_flight",
				Help: "Number of HTTP requests currently executing, excluding metrics scrapes.",
			},
		),
	}

	m.registry.MustRegister(
		m.requests,
		m.duration,
		m.inFlight,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(
		m.registry,
		promhttp.HandlerOpts{},
	)
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Collecting metrics should not inflate application traffic.
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		m.inFlight.Inc()
		defer m.inFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		method := c.Request.Method
		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
			// Keep standard methods.
		default:
			method = "OTHER"
		}

		status := strconv.Itoa(c.Writer.Status())

		m.requests.WithLabelValues(method, route, status).Inc()
		m.duration.WithLabelValues(method, route).
			Observe(time.Since(start).Seconds())
	}
}
