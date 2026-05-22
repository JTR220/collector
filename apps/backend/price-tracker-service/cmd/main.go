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
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/JTR220/collector/price-tracker-service/config"
	"github.com/JTR220/collector/price-tracker-service/internal/consumer"
	"github.com/JTR220/collector/price-tracker-service/internal/detector"
	"github.com/JTR220/collector/price-tracker-service/internal/handler"
	"github.com/JTR220/collector/price-tracker-service/internal/repository"
)

func main() {
	// ── Logger ───────────────────────────────────────────────
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting price-tracker-service")

	// ── Config ───────────────────────────────────────────────
	cfg := config.Load()

	// ── Database ─────────────────────────────────────────────
	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to postgres")
	}
	defer db.Close()
	log.Info().Msg("connected to postgres")

	repo := repository.NewPriceRepository(db)
	if err := repo.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	// ── RabbitMQ ─────────────────────────────────────────────
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to rabbitmq")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open rabbitmq channel")
	}
	defer ch.Close()

	if err := consumer.Setup(ch, &cfg.RabbitMQ); err != nil {
		log.Fatal().Err(err).Msg("rabbitmq setup failed")
	}
	log.Info().Msg("connected to rabbitmq")

	// ── Components ───────────────────────────────────────────
	pub := consumer.NewPublisher(ch, cfg.RabbitMQ.ExchangeAlerts)
	det := detector.New(repo, cfg.Rules)
	priceConsumer := consumer.NewPriceConsumer(repo, det, pub, cfg)

	// ── HTTP Server ──────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	h := handler.New(repo)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// ── Graceful shutdown ────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start RabbitMQ consumer in background
	go func() {
		if err := priceConsumer.Start(ctx, ch); err != nil {
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

	log.Info().Msg("price-tracker-service stopped")
}
