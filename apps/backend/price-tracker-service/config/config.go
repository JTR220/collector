package config

import (
	"os"
	"strconv"

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
	Rules    DetectionRules
}

type ServerConfig struct {
	Port string `mapstructure:"PORT"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"DATABASE_DSN"`
}

type RabbitMQConfig struct {
	URL              string `mapstructure:"RABBITMQ_URL"`
	ExchangeEvents   string `mapstructure:"RABBITMQ_EXCHANGE_EVENTS"`
	ExchangeAlerts   string `mapstructure:"RABBITMQ_EXCHANGE_ALERTS"`
	QueuePriceUpdate string `mapstructure:"RABBITMQ_QUEUE_PRICE_UPDATE"`
}

type DetectionRules struct {
	SpikeThresholdPercent float64 `mapstructure:"SPIKE_THRESHOLD_PERCENT"`
	SpikeWindowHours      int     `mapstructure:"SPIKE_WINDOW_HOURS"`
	FloodMaxUpdates       int     `mapstructure:"FLOOD_MAX_UPDATES"`
	FloodWindowMinutes    int     `mapstructure:"FLOOD_WINDOW_MINUTES"`
}

func Load() *Config {
	viper.SetDefault("PORT", "8082")
	viper.SetDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/collector_price?sslmode=disable")
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("RABBITMQ_EXCHANGE_EVENTS", "collector.events")
	viper.SetDefault("RABBITMQ_EXCHANGE_ALERTS", "collector.alerts")
	viper.SetDefault("RABBITMQ_QUEUE_PRICE_UPDATE", "price-tracker.price.updated")
	viper.SetDefault("SPIKE_THRESHOLD_PERCENT", 50.0)
	viper.SetDefault("SPIKE_WINDOW_HOURS", 24)
	viper.SetDefault("FLOOD_MAX_UPDATES", 5)
	viper.SetDefault("FLOOD_WINDOW_MINUTES", 60)

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
	cfg.RabbitMQ.QueuePriceUpdate = envOr("RABBITMQ_QUEUE_PRICE_UPDATE", cfg.RabbitMQ.QueuePriceUpdate)
	if v, err := strconv.ParseFloat(os.Getenv("SPIKE_THRESHOLD_PERCENT"), 64); err == nil {
		cfg.Rules.SpikeThresholdPercent = v
	}
	if v, err := strconv.Atoi(os.Getenv("SPIKE_WINDOW_HOURS")); err == nil {
		cfg.Rules.SpikeWindowHours = v
	}
	if v, err := strconv.Atoi(os.Getenv("FLOOD_MAX_UPDATES")); err == nil {
		cfg.Rules.FloodMaxUpdates = v
	}
	if v, err := strconv.Atoi(os.Getenv("FLOOD_WINDOW_MINUTES")); err == nil {
		cfg.Rules.FloodWindowMinutes = v
	}

	return &cfg
}
