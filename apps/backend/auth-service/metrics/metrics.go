package metrics

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serviceName = "auth-service"

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

	LoginAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_auth_login_attempts_total",
		Help: "Authentication login attempts by result.",
	}, []string{"result"})
	RegistrationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_auth_registrations_total",
		Help: "User registration attempts by result.",
	}, []string{"result"})
	JWTRejectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_auth_jwt_rejections_total",
		Help: "JWT authentication rejections by reason.",
	}, []string{"reason"})
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

func RecordLogin(result string) {
	LoginAttemptsTotal.WithLabelValues(result).Inc()
}

func RecordRegistration(result string) {
	RegistrationsTotal.WithLabelValues(result).Inc()
}

func RecordJWTRejection(reason string) {
	JWTRejectionsTotal.WithLabelValues(reason).Inc()
}

// Serve expose /metrics sur un port interne dedie (jamais route par l'ingress
// public, qui ne proxy que le port "http" du Service k8s) : le scraping
// Prometheus reste possible en cluster sans exposer de metriques metier
// (tentatives de login, etc.) sur Internet.
func Serve(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Le serveur de metriques n'a pas pu demarrer : %v", err)
	}
}
