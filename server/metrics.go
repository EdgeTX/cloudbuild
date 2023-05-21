package server

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const ()

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

	durationBuckets        = []float64{100, 200, 400, 800, 2000}
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
	metricBuildRequestQueued = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_request_queued",
			Help: "Number of build requests queued.",
		},
		[]string{"release", "target"},
	)
	metricBuildRequestBuilding = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_request_building",
			Help: "Number of build requests building.",
		},
		[]string{"release", "target"},
	)
	metricBuildRequestFailed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_request_failed",
			Help: "Number of build requests failed.",
		},
		[]string{"release", "target"},
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
		metricRequestBody.Add(float64(w.Size()))
	}

	latency := time.Since(start)
	metricsRequestDuration.Observe(float64(latency.Milliseconds()))
}

func RegisterMetrics() {
	prometheus.MustRegister(
		metricRequestTotal,
		metricRequestBody,
		metricResponseBody,
		metricsRequestDuration,
		metricBuildRequestTotal,
		metricBuildRequestQueued,
		metricBuildRequestBuilding,
		metricBuildRequestFailed,
	)
}
