package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

type Server struct {
	mux        *http.ServeMux
	config     Config
	logger     logger.Logger
	middleware []Middleware
	server     *http.Server
}

func NewServer(
	config Config,
	log logger.Logger,
	middleware ...Middleware,
) *Server {
	mux := http.NewServeMux()

	return &Server{
		mux:        mux,
		config:     config,
		logger:     log,
		middleware: middleware,
		server: &http.Server{
			Addr:              config.Addr(),
			Handler:           mux,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

func (s *Server) Name() string {
	return "http"
}

func (s *Server) RegisterRouter(router *Router) {
	for _, route := range router.routes {
		pattern := string(route.Method) +
			" /api/" +
			string(router.apiVersion) +
			route.Path

		middleware := make(
			[]Middleware,
			0,
			len(s.middleware)+len(router.middleware),
		)

		middleware = append(middleware, s.middleware...)
		middleware = append(middleware, router.middleware...)

		handler := route.HandlerFunc(middleware...)

		s.mux.Handle(pattern, s.wrap(handler))
	}
}

// wrap — единственное место HTTP-транспорта, где известен базовый
// логгер. Здесь он обогащается полями запроса и передаётся в Context.
func (s *Server) wrap(handler HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := s.logger.With(
			logger.String("request_id", uuid.NewString()),
			logger.String("transport", "http"),
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("remote_addr", r.RemoteAddr),
		)

		ctx := newContext(w, r, log)
		handler(ctx)
	})
}

// Start запускает HTTP-сервер в фоне. Контекст используется для
// прерывания установки listener'а; сам сервер останавливается
// через Shutdown.
func (s *Server) Start(ctx context.Context) error {
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", s.config.Addr())
	if err != nil {
		return fmt.Errorf("listen HTTP server: %w", err)
	}

	go func() {
		if err := s.server.Serve(listener); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(
				"HTTP server failed",
				logger.Error(err),
			)
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	return nil
}
