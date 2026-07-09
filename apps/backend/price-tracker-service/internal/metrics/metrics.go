package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const serviceName = "price-tracker-service"

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_http_requests_total",
		Help: "Total HTTP requests handled by Collector services.",
	}, []string{"service", "method", "route", "status"})
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "collector_http_request_duration_seconds",
		Help:    "HTTP request duration by route and status.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "route", "status"})
	HTTPInFlightRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "collector_http_in_flight_requests",
		Help: "Current in-flight HTTP requests by Collector service.",
	}, []string{"service"})

	PriceEventsConsumedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_price_events_consumed_total",
		Help: "price.updated events consumed by result.",
	}, []string{"result"})
	FraudAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_price_fraud_alerts_total",
		Help: "Fraud alerts detected by reason.",
	}, []string{"reason"})
	RabbitMQErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_price_rabbitmq_errors_total",
		Help: "RabbitMQ errors by operation.",
	}, []string{"operation"})
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		HTTPInFlightRequests.WithLabelValues(serviceName).Inc()
		defer HTTPInFlightRequests.WithLabelValues(serviceName).Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		HTTPRequestsTotal.WithLabelValues(serviceName, c.Request.Method, route, status).Inc()
		HTTPRequestDuration.WithLabelValues(serviceName, c.Request.Method, route, status).Observe(time.Since(start).Seconds())
	}
}

func RecordPriceEvent(result string) {
	PriceEventsConsumedTotal.WithLabelValues(result).Inc()
}

func RecordFraudAlert(reason string) {
	FraudAlertsTotal.WithLabelValues(reason).Inc()
}

func RecordRabbitMQError(operation string) {
	RabbitMQErrorsTotal.WithLabelValues(operation).Inc()
}
