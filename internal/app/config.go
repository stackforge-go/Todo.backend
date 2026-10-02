package app

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ShutdownTimeout time.Duration  `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	TimeZone        *time.Location `envconfig:"TIME_ZONE" default:"UTC"`
	Env             string         `envconfig:"ENV" default:"production"`
	CookieDomain    string         `envconfig:"COOKIE_DOMAIN" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("APP", &config); err != nil {
		return Config{}, fmt.Errorf("process env config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("process Application config: %w", err)
		panic(err)
	}

	return config
}
