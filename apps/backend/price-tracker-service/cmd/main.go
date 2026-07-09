package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // enregistre le driver "postgres" pour database/sql (utilise via sqlx.Connect)
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/consumer"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/handler"
	"github.com/JTR220/collector/price-tracker-service/internal/metrics"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// ── Logger ───────────────────────────────────────────────
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting price-tracker-service")

	// ── Config ───────────────────────────────────────────────
	cfg := config.Load()

	// Fail-fast : aucun fallback de secret dans le code (docker-compose ou
	// Sealed Secret k8s doivent fournir JWT_SECRET).
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal().Msg("JWT_SECRET est requis : definissez-le dans l'environnement")
	}

	// ── Database ─────────────────────────────────────────────
	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to postgres")
	}
	defer func() { _ = db.Close() }()
	// Borne le pool de connexions (PostgreSQL est partage entre les services).
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	log.Info().Msg("connected to postgres")

	repo := repository.NewPriceRepository(db)
	if err := repo.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	// ── RabbitMQ ─────────────────────────────────────────────
	// Verification de connectivite au demarrage (fail-fast) : la connexion
	// reelle et sa reconnexion automatique sont gerees par PriceConsumer.Start.
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to rabbitmq")
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		log.Fatal().Err(err).Msg("cannot open rabbitmq channel")
	}
	if err := consumer.Setup(ch, &cfg.RabbitMQ); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		log.Fatal().Err(err).Msg("rabbitmq setup failed")
	}
	_ = ch.Close()
	_ = conn.Close()
	log.Info().Msg("connected to rabbitmq")

	// ── Components ───────────────────────────────────────────
	pub := consumer.NewPublisher(nil, cfg.RabbitMQ.ExchangeAlerts)
	det := detector.New(repo, cfg.Rules)
	priceConsumer := consumer.NewPriceConsumer(repo, det, pub, cfg)

	// ── HTTP Server ──────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(metrics.Middleware())

	h := handler.New(repo)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// /metrics sur un port interne dedie, jamais route par l'ingress public
	// (qui ne proxy que le port "http" du Service k8s) : evite d'exposer des
	// metriques metier (alertes fraude, evenements prix...) sur Internet.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:    ":9100",
		Handler: metricsMux,
	}

	// ── Graceful shutdown ────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start RabbitMQ consumer in background — gere sa propre reconnexion.
	go func() {
		if err := priceConsumer.Start(ctx, cfg.RabbitMQ.URL); err != nil {
			log.Error().Err(err).Msg("consumer error")
			cancel()
		}
	}()

	// Start HTTP server in background
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	go func() {
		log.Info().Str("port", "9100").Msg("metrics server listening")
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("metrics server error")
		}
	}()

	// Wait for OS signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("metrics server shutdown error")
	}

	log.Info().Msg("price-tracker-service stopped")
}

// corsMiddleware reprend la convention du catalog-service (FRONTEND_ORIGIN).
func corsMiddleware() gin.HandlerFunc {
	allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
