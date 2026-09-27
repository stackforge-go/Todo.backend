package http

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host           string   `envconfig:"HOST" required:"true"`
	Port           int      `envconfig:"PORT" required:"true"`
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" required:"true"`
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("HTTP", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("process HTTP config: %w", err))
	}

	return config
}
