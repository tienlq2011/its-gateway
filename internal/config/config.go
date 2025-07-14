package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"its-gateway/internal/mq"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	RabbitMQ mq.Config      `yaml:"rabbitmq"`
	Dahua    DahuaConfig    `yaml:"dahua"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DahuaConfig struct {
	Username string            `yaml:"username"`
	Password string            `yaml:"password"`
	LaneMap  map[string]string `yaml:"laneMap"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
