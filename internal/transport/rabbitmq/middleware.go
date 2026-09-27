package rabbitmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

type Middleware func(next HandlerFunc) HandlerFunc

func ChainMiddleware(h HandlerFunc, m ...Middleware) HandlerFunc {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func LoggingMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

			err := next(ctx)

			fields := []logger.Field{
				logger.String("routing_key", ctx.RoutingKey()),
				logger.String("message_id", ctx.MessageID()),
				logger.Duration("duration", time.Since(start)),
			}

			switch {
			case err == nil:
				ctx.Logger().Info("message handled", fields...)
			case errors.Is(err, ErrRequeue):
				ctx.Logger().Warn("message requeued", append(fields, logger.Error(err))...)
			case errors.Is(err, ErrDrop):
				ctx.Logger().Warn("message dropped", append(fields, logger.Error(err))...)
			default:
				ctx.Logger().Error("message failed", append(fields, logger.Error(err))...)
			}

			return err
		}
	}
}

func RecoverMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					err = errs.Internal.WithMessage("panic recovered").WithOp("rabbitmq.recover").Wrap(fmt.Errorf("panic: %v", rec))
				}
			}()
			return next(ctx)
		}
	}
}
