package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
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
	viper.SetDefault("JWT_SECRET", "change-me-in-production")

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return &cfg
}
