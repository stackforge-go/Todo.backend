package goredis

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config — конфигурация Redis-клиента.
// Все поля читаются из переменных окружения с префиксом REDIS_.
type config struct {
	Host     string        `envconfig:"HOST" required:"true"`
	Port     string        `envconfig:"PORT" default:"6379"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	Database int           `envconfig:"DB" default:"0"`
	TTL      time.Duration `envconfig:"TTL" default:"5m"`
}

// NewConfig читает конфигурацию сервера из переменных окружения.
func NewConfig() (config, error) {
	var cfg config

	if err := envconfig.Process("REDIS", &cfg); err != nil {
		return config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return cfg, nil
}

// NewConfigMust — «Must»-вариант конструктора: паникует при ошибке.
func NewConfigMust() config {
	cfg, err := NewConfig()
	if err != nil {
		err := fmt.Errorf("get Redis pool config: %w", err)
		panic(err)
	}

	return cfg
}

func (c config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
