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

	"github.com/JTR220/collector/notification-service/config"
	"github.com/JTR220/collector/notification-service/internal/authclient"
	"github.com/JTR220/collector/notification-service/internal/consumer"
	"github.com/JTR220/collector/notification-service/internal/handler"
	"github.com/JTR220/collector/notification-service/internal/hub"
	"github.com/JTR220/collector/notification-service/internal/mailer"
	"github.com/JTR220/collector/notification-service/internal/metrics"
	"github.com/JTR220/collector/notification-service/internal/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting notification-service")

	cfg := config.Load()

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

	repo := repository.New(db)
	if err := repo.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}
	if err := repo.SeedDemoData(context.Background()); err != nil {
		log.Warn().Err(err).Msg("seed demo data failed (non-fatal)")
	}

	// Verification de connectivite au demarrage (fail-fast) : la connexion
	// reelle et sa reconnexion automatique sont gerees par Manager.Start.
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

	wsHub := hub.New()
	go wsHub.Run()

	mail := mailer.New(mailer.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		From:     cfg.SMTP.From,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
	})
	authCli := authclient.New(cfg.Internal.AuthServiceURL, cfg.Internal.Secret)

	mgr := consumer.NewManager(wsHub, repo, cfg, mail, authCli)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(metrics.Middleware())

	h := handler.New(wsHub, repo, cfg.JWT.Secret, authCli, cfg.Internal.Secret)
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// /metrics sur un port interne dedie, jamais route par l'ingress public
	// (qui ne proxy que le port "http" du Service k8s) : evite d'exposer des
	// metriques metier (messages, emails, erreurs RabbitMQ...) sur Internet.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:              ":9100",
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mgr.Start(ctx, cfg.RabbitMQ.URL)
	go runRetentionWorker(ctx, repo, cfg.Retention.Days)

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

	log.Info().Msg("notification-service stopped")
}

// runRetentionWorker purge periodiquement les notifications et messages
// au-dela de la duree de conservation configuree (RETENTION_DAYS, 365 jours
// par defaut) : minimisation des donnees (art. 5.1.e RGPD). Une premiere
// purge tourne au demarrage, puis toutes les 24h jusqu'a l'arret du service.
func runRetentionWorker(ctx context.Context, repo *repository.NotificationRepository, retentionDays int) {
	purge := func() {
		cutoff := time.Now().AddDate(0, 0, -retentionDays)
		notifs, messages, err := repo.PurgeOlderThan(ctx, cutoff)
		if err != nil {
			log.Error().Err(err).Msg("purge de retention echouee")
			return
		}
		if notifs > 0 || messages > 0 {
			log.Info().Int64("notifications", notifs).Int64("messages", messages).
				Msg("purge de retention effectuee")
		}
	}

	purge()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			purge()
		}
	}
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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		// Le cookie de session httpOnly doit transiter sur les requetes
		// cross-origin front -> API : Allow-Credentials cote serveur +
		// credentials:'include' cote client (fetch).
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
