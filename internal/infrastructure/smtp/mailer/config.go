package mailer

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     int    `envconfig:"PORT" required:"true"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	From     string `envconfig:"FROM" required:"true"`
}

func NewConfig() (config, error) {
	var cfg config
	if err := envconfig.Process("SMTP", &cfg); err != nil {
		return config{}, fmt.Errorf("envconfig process: %w", err)
	}

	return cfg, nil
}

func NewConfigMust() config {
	cfg, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("process SMTP config: %w", err)
		panic(err)
	}

	return cfg
}
