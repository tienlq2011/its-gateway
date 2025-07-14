package mq

import "time"

// Config holds RabbitMQ connection and publishing configuration
type Config struct {
	URL          string        `yaml:"url"`
	Exchange     string        `yaml:"exchange"`
	ExchangeType string        `yaml:"exchangeType"`
	RoutingKey   string        `yaml:"routingKey"`
	Retry        int           `yaml:"retry"`
	RetryDelay   time.Duration `yaml:"retryDelay"`
}
