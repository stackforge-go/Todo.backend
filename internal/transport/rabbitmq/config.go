package rabbitmq

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     int    `envconfig:"PORT" required:"true"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	VHost    string `envconfig:"VHOST" default:"/"`

	PrefetchCount  int           `envconfig:"PREFETCH_COUNT" default:"10"`
	ReconnectDelay time.Duration `envconfig:"RECONNECT_DELAY" default:"5s"`
}

func NewConfig() (Config, error) {
	var cfg Config

	if err := envconfig.Process("RMQ", &cfg); err != nil {
		return Config{}, fmt.Errorf("process enconfig: %w", err)
	}

	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("process RMQ config: %w", err)
		panic(err)
	}

	return cfg
}

func (c *Config) URL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.VHost,
	)
}
