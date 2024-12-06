package main

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	HostPort        string `env:"HOST_PORT, required"`
	KeyCloakCertURL string `env:"KEYCLOAK_CERT_URL, required"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process(context.Background(), &cfg)
	if err != nil {
		return nil, fmt.Errorf("env processing: %w", err)
	}
	return &cfg, nil
}
