package server

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTP metrics.
	metricRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of requests.",
		},
		[]string{"code"},
	)

	metricRequestBody = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_request_body_total",
		Help: "Total size of request bodies.",
	})

	metricResponseBody = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_response_body_total",
		Help: "Total size of response bodies.",
	})

	durationBuckets        = []float64{10, 50, 200, 500}
	metricsRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration",
		Help:    "Request processing duration.",
		Buckets: durationBuckets,
	})
	// Firmware metrics.
	metricBuildRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_request_total",
			Help: "Total number of build requests.",
		},
		[]string{"release", "target"},
	)
	metricBuildRequestQueued = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "build_request_queued",
			Help: "Number of build requests queued.",
		},
	)
	metricBuildRequestBuilding = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "build_request_building",
			Help: "Number of build requests building.",
		},
	)
	metricBuildRequestFailed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "build_request_failed",
			Help: "Number of build requests failed.",
		},
	)
)

func GinMetrics(c *gin.Context) {
	startTime := time.Now()
	c.Next()
	ginMetricsHandle(c, startTime)
}

func ginMetricsHandle(c *gin.Context, start time.Time) {
	r := c.Request
	w := c.Writer

	code := strconv.Itoa(w.Status())
	metricRequestTotal.WithLabelValues(code).Inc()

	if r.ContentLength > 0 {
		metricRequestBody.Add(float64(r.ContentLength))
	}

	if w.Size() > 0 {
		metricResponseBody.Add(float64(w.Size()))
	}

	latency := float64(time.Since(start).Milliseconds())
	metricsRequestDuration.Observe(latency)
}

func RegisterMetrics() *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(
		metricRequestTotal,
		metricRequestBody,
		metricResponseBody,
		metricsRequestDuration,
		metricBuildRequestTotal,
		metricBuildRequestQueued,
		metricBuildRequestBuilding,
		metricBuildRequestFailed,
	)
	return r
}
