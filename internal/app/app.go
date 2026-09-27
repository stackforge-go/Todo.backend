package app

import (
	"context"
	"fmt"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

// Component — единица приложения, имеющая собственный жизненный цикл.
//
// Start получает контекст приложения, который живёт до shutdown.
// Всё, что компонент запускает в фоне (горутины, поллеры, long-poll),
// должно наследовать этот контекст, чтобы корректно остановиться.
//
// Shutdown получает контекст с таймаутом; компонент обязан уложиться
// в него либо вернуть ошибку.
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type App struct {
	config     Config
	logger     logger.Logger
	components []Component
}

func NewApp(
	config Config,
	log logger.Logger,
	components ...Component,
) *App {
	return &App{
		config:     config,
		logger:     log,
		components: components,
	}
}

func (a *App) Start(ctx context.Context) error {
	for _, component := range a.components {
		if err := component.Start(ctx); err != nil {
			a.logger.Error(
				"application component failed",
				logger.String("component", component.Name()),
				logger.Error(err),
			)

			return fmt.Errorf(
				"start component %q: %w",
				component.Name(),
				err,
			)
		}

		a.logger.Info(
			"application component started",
			logger.String("component", component.Name()),
		)
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	var shutdownErr error

	// Останавливаем компоненты в обратном порядке запуска.
	for i := len(a.components) - 1; i >= 0; i-- {
		component := a.components[i]

		if err := component.Shutdown(ctx); err != nil {
			a.logger.Error(
				"application component shutdown failed",
				logger.String("component", component.Name()),
				logger.Error(err),
			)

			if shutdownErr == nil {
				shutdownErr = fmt.Errorf(
					"shutdown component %q: %w",
					component.Name(),
					err,
				)
			}

			continue
		}

		a.logger.Info(
			"application component stopped",
			logger.String("component", component.Name()),
		)
	}

	return shutdownErr
}
