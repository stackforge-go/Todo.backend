package golangjwt

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Secret     string        `envconfig:"SECRET" required:"true"`
	AccessTTL  time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"REFRESH_TTL" default:"168h"`
	Issuer     string        `envconfig:"ISSUER" default:"todo-backend"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("JWT", &cfg); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("jwt config: %w", err)
		panic(err)
	}
	return cfg
}
