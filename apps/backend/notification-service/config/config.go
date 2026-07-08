package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// envOr renvoie la variable d'environnement si elle est définie, sinon def.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	Internal InternalConfig
}

type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     string `mapstructure:"SMTP_PORT"`
	From     string `mapstructure:"SMTP_FROM"`
	User     string `mapstructure:"SMTP_USER"`
	Password string `mapstructure:"SMTP_PASSWORD"`
}

// InternalConfig porte les coordonnees d'appel inter-services (resolution de
// l'email du vendeur/acheteur aupres d'auth-service pour l'envoi d'email).
type InternalConfig struct {
	AuthServiceURL string `mapstructure:"AUTH_SERVICE_URL"`
	Secret         string `mapstructure:"INTERNAL_SECRET"`
}

type ServerConfig struct {
	Port string `mapstructure:"PORT"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"DATABASE_DSN"`
}

type RabbitMQConfig struct {
	URL             string `mapstructure:"RABBITMQ_URL"`
	ExchangeEvents  string `mapstructure:"RABBITMQ_EXCHANGE_EVENTS"`
	ExchangeAlerts  string `mapstructure:"RABBITMQ_EXCHANGE_ALERTS"`
	QueuePriceNotif string `mapstructure:"RABBITMQ_QUEUE_PRICE_NOTIF"`
	QueueFraudNotif string `mapstructure:"RABBITMQ_QUEUE_FRAUD_NOTIF"`
}

type JWTConfig struct {
	Secret string `mapstructure:"JWT_SECRET"`
}

func Load() *Config {
	viper.SetDefault("PORT", "8083")
	viper.SetDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/collector_notifications?sslmode=disable")
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("RABBITMQ_EXCHANGE_EVENTS", "collector.events")
	viper.SetDefault("RABBITMQ_EXCHANGE_ALERTS", "collector.alerts")
	viper.SetDefault("RABBITMQ_QUEUE_PRICE_NOTIF", "notification-service.price.updated")
	viper.SetDefault("RABBITMQ_QUEUE_FRAUD_NOTIF", "notification-service.fraud.alert")
	viper.SetDefault("SMTP_PORT", "1025")
	viper.SetDefault("SMTP_FROM", "notifications@collector.shop")
	viper.SetDefault("AUTH_SERVICE_URL", "http://localhost:8080")

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// viper.AutomaticEnv() n'alimente pas Unmarshal de façon fiable : on applique
	// explicitement les surcharges d'environnement par-dessus les valeurs par défaut.
	cfg.Server.Port = envOr("PORT", cfg.Server.Port)
	cfg.Database.DSN = envOr("DATABASE_DSN", cfg.Database.DSN)
	cfg.RabbitMQ.URL = envOr("RABBITMQ_URL", cfg.RabbitMQ.URL)
	cfg.RabbitMQ.ExchangeEvents = envOr("RABBITMQ_EXCHANGE_EVENTS", cfg.RabbitMQ.ExchangeEvents)
	cfg.RabbitMQ.ExchangeAlerts = envOr("RABBITMQ_EXCHANGE_ALERTS", cfg.RabbitMQ.ExchangeAlerts)
	cfg.RabbitMQ.QueuePriceNotif = envOr("RABBITMQ_QUEUE_PRICE_NOTIF", cfg.RabbitMQ.QueuePriceNotif)
	cfg.RabbitMQ.QueueFraudNotif = envOr("RABBITMQ_QUEUE_FRAUD_NOTIF", cfg.RabbitMQ.QueueFraudNotif)
	cfg.JWT.Secret = envOr("JWT_SECRET", cfg.JWT.Secret)
	cfg.SMTP.Host = envOr("SMTP_HOST", cfg.SMTP.Host)
	cfg.SMTP.Port = envOr("SMTP_PORT", cfg.SMTP.Port)
	cfg.SMTP.From = envOr("SMTP_FROM", cfg.SMTP.From)
	cfg.SMTP.User = envOr("SMTP_USER", cfg.SMTP.User)
	cfg.SMTP.Password = envOr("SMTP_PASSWORD", cfg.SMTP.Password)
	cfg.Internal.AuthServiceURL = envOr("AUTH_SERVICE_URL", cfg.Internal.AuthServiceURL)
	cfg.Internal.Secret = envOr("INTERNAL_SECRET", cfg.Internal.Secret)

	// Fail-fast : aucun secret par defaut dans le code. docker-compose ou le
	// Sealed Secret k8s doivent fournir JWT_SECRET (il doit etre identique a
	// celui d'auth-service, qui signe les tokens).
	if cfg.JWT.Secret == "" {
		log.Fatal().Msg("JWT_SECRET est requis : definissez-le dans l'environnement")
	}

	return &cfg
}
