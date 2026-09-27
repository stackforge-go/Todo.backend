package http

import (
	"fmt"
	"time"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

type Middleware func(next HandlerFunc) HandlerFunc

func ChainMiddleware(
	handler HandlerFunc,
	middleware ...Middleware,
) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	return handler
}

// LoggingMiddleware пишет одну запись на запрос — со статусом,
// длительностью, размером ответа и, если была, доменной ошибкой.
func LoggingMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) {
			start := time.Now()

			next(ctx)

			log := ctx.Logger()
			status := ctx.Response().Status()

			fields := []logger.Field{
				logger.Int("status", status),
				logger.Int("bytes", ctx.Response().BytesWritten()),
				logger.Duration("duration", time.Since(start)),
			}

			if appErr := ctx.AppError(); appErr != nil {
				fields = append(fields,
					logger.String("code", string(appErr.Code)),
					logger.String("op", appErr.Op),
					logger.Error(appErr),
				)
			}

			switch {
			case status >= 500:
				log.Error("HTTP request", fields...)
			case status >= 400:
				log.Warn("HTTP request", fields...)
			default:
				log.Info("HTTP request", fields...)
			}
		}
	}
}

func RecoverMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) {
			defer func() {
				if rec := recover(); rec != nil {
					err := errs.Internal.
						WithMessage("internal error").
						WithOp("http.recover").
						Wrap(fmt.Errorf("panic: %v", rec))

					ctx.Error(err)
				}
			}()

			next(ctx)
		}
	}
}
