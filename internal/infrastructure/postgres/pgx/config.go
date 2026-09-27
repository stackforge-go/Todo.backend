package pgx

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config хранит параметры подключения к PostgreSQL.
// Все поля читаются из переменных окружения с префиксом "POSTGRES_":
// POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, POSTGRES_TIMEOUT.
type config struct {
	Host     string        `envconfig:"HOST" required:"true"`
	Port     string        `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER" required:"true"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	Database string        `envconfig:"DB" required:"true"`
	Timeout  time.Duration `envconfig:"TIMEOUT" required:"true"`
}

// NewConfig читает конфигурацию сервера из переменных окружения.
func NewConfig() (config, error) {
	var cfg config
	if err := envconfig.Process("POSTGRES", &cfg); err != nil {
		return config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return cfg, nil
}

// NewConfigMust — «Must»-вариант конструктора: паникует при ошибке.
func NewConfigMust() config {
	cfg, err := NewConfig()
	if err != nil {
		err := fmt.Errorf("get Postgres connection pool config: %w", err)
		panic(err)
	}

	return cfg
}
