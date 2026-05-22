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

	"github.com/JTR220/collector/notification-service/config"
	"github.com/JTR220/collector/notification-service/internal/consumer"
	"github.com/JTR220/collector/notification-service/internal/handler"
	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/repository"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting notification-service")

	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to postgres")
	}
	defer db.Close()
	log.Info().Msg("connected to postgres")

	repo := repository.New(db)
	if err := repo.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

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

	wsHub := hub.New()
	go wsHub.Run()

	mgr := consumer.NewManager(ch, wsHub, repo, cfg)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	h := handler.New(wsHub, repo, cfg.JWT.Secret)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mgr.Start(ctx)

	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

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

	log.Info().Msg("notification-service stopped")
}
