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

const serviceName = "catalog-service"

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

	ArticlesCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_catalog_articles_created_total",
		Help: "Article creation attempts by result.",
	}, []string{"result"})
	OrdersCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_catalog_orders_created_total",
		Help: "Marketplace order creation attempts by result.",
	}, []string{"result"})
	OrderDecisionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_catalog_order_decisions_total",
		Help: "Seller order decisions by decision and result.",
	}, []string{"decision", "result"})
	ImageUploadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_catalog_image_uploads_total",
		Help: "Article image upload attempts by result.",
	}, []string{"result"})
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

func RecordArticleCreated(result string) {
	ArticlesCreatedTotal.WithLabelValues(result).Inc()
}

func RecordOrderCreated(result string) {
	OrdersCreatedTotal.WithLabelValues(result).Inc()
}

func RecordOrderDecision(decision, result string) {
	OrderDecisionsTotal.WithLabelValues(decision, result).Inc()
}

func RecordImageUpload(result string) {
	ImageUploadsTotal.WithLabelValues(result).Inc()
}

// Serve expose /metrics sur un port interne dedie (jamais route par l'ingress
// public, qui ne proxy que le port "http" du Service k8s) : le scraping
// Prometheus reste possible en cluster sans exposer de metriques metier
// sur Internet.
func Serve(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Le serveur de metriques n'a pas pu demarrer : %v", err)
	}
}
