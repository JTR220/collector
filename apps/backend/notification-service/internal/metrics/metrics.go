package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const serviceName = "notification-service"

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

	NotificationsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_notifications_created_total",
		Help: "Notifications created by type and result.",
	}, []string{"type", "result"})
	MessagesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_messages_sent_total",
		Help: "Direct messages sent by result.",
	}, []string{"result"})
	EmailsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_emails_sent_total",
		Help: "Transactional emails by result.",
	}, []string{"result"})
	WebSocketActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "collector_websocket_active_connections",
		Help: "Current active WebSocket connections.",
	})
	RabbitMQErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_notification_rabbitmq_errors_total",
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

// inc factorise l'increment des compteurs metier a une ou plusieurs
// dimensions (type/result/operation...), evitant une fonction Record*
// quasi identique par compteur.
func inc(counter *prometheus.CounterVec, labelValues ...string) {
	counter.WithLabelValues(labelValues...).Inc()
}

func RecordNotification(notificationType, result string) {
	inc(NotificationsCreatedTotal, notificationType, result)
}
func RecordMessage(result string) { inc(MessagesSentTotal, result) }
func RecordEmail(result string)   { inc(EmailsSentTotal, result) }

func SetWebSocketActiveConnections(count int) {
	WebSocketActiveConnections.Set(float64(count))
}

func RecordRabbitMQError(operation string) { inc(RabbitMQErrorsTotal, operation) }
