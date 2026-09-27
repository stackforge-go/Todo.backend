package zap

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Level  string `envconfig:"LEVEL" default:"DEBUG"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

func NewConfig() (config, error) {
	var cfg config

	if err := envconfig.Process("LOGGER", &cfg); err != nil {
		return config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return cfg, nil
}

func NewConfigMust() config {
	cfg, err := NewConfig()
	if err != nil {
		err := fmt.Errorf("get Zap Logger config: %w", err)
		panic(err)
	}

	return cfg
}
