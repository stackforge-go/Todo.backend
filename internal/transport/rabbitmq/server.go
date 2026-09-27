package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

type Server struct {
	config     Config
	logger     logger.Logger
	middleware []Middleware
	routes     []Route

	mu   sync.Mutex
	conn *amqp.Connection
	stop context.CancelFunc
	wg   sync.WaitGroup
}

func NewServer(config Config, log logger.Logger, middleware ...Middleware) (*Server, error) {
	conn, err := dialWithRetry(config.URL(), 10, log)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	return &Server{
		config:     config,
		logger:     log,
		middleware: middleware,
		conn:       conn,
	}, nil
}

func dialWithRetry(url string, attempts int, log logger.Logger) (*amqp.Connection, error) {
	var lastErr error

	for i := 0; i < attempts; i++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		delay := time.Duration(i+1) * time.Second

		log.Warn("rabbitmq dial failed, retrying",
			logger.Int("attempt", i+1),
			logger.Duration("delay", delay),
			logger.Error(err),
		)

		time.Sleep(delay)
	}

	return nil, fmt.Errorf("after %d attempts: %w", attempts, lastErr)
}

func (s *Server) Name() string { return "rabbitmq" }

func (s *Server) RegisterRouter(router *Router) {
	for _, route := range router.routes {
		s.routes = append(s.routes, route)
	}
}

func (s *Server) Publisher() (*Publisher, error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open publisher channel: %w", err)
	}
	return NewPublisher(ch), nil
}

func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	startCtx, cancel := context.WithCancel(ctx)
	s.stop = cancel
	s.mu.Unlock()

	for _, route := range s.routes {
		r := route
		concurrency := r.Concurrency
		if concurrency <= 0 {
			concurrency = 1
		}

		for i := 0; i < concurrency; i++ {
			s.wg.Add(1)
			go func(workerID int) {
				defer s.wg.Done()
				s.consumeLoop(startCtx, r, workerID)
			}(i)
		}
	}

	return nil
}

func (s *Server) consumeLoop(ctx context.Context, route Route, workerID int) {
	log := s.logger.With(
		logger.String("queue", route.Queue),
		logger.Int("worker_id", workerID),
	)

	for {
		if ctx.Err() != nil {
			return
		}

		err := s.consume(ctx, route, log)
		if err == nil {
			return
		}

		log.Error("consumer stopped, reconnecting",
			logger.Error(err),
			logger.Duration("delay", s.config.ReconnectDelay),
		)

		select {
		case <-ctx.Done():
			return
		case <-time.After(s.config.ReconnectDelay):
		}
	}
}

func (s *Server) consume(ctx context.Context, route Route, log logger.Logger) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	prefetch := route.Prefetch
	if prefetch <= 0 {
		prefetch = s.config.PrefetchCount
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	if _, err := ch.QueueDeclare(
		route.Queue,
		true, false, false, false,
		route.QueueArgs,
	); err != nil {
		return fmt.Errorf("declare queue %q: %w", route.Queue, err)
	}

	msgs, err := ch.Consume(
		route.Queue,
		"", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	handler := ChainMiddleware(route.Handler, append(s.middleware, route.Middleware...)...)

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return errors.New("delivery channel closed")
			}
			s.handle(ctx, route, d, ch, handler, log)
		}
	}
}

func (s *Server) handle(
	parentCtx context.Context,
	route Route,
	d amqp.Delivery,
	ch *amqp.Channel,
	handler HandlerFunc,
	log logger.Logger,
) {
	msgLog := log.With(
		logger.String("message_id", d.MessageId),
		logger.String("routing_key", d.RoutingKey),
		logger.String("content_type", d.ContentType),
	)

	msgCtx := newContext(parentCtx, d, ch, msgLog)

	var err error
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("panic: %v", rec)
				msgLog.Error("panic in handler", logger.Any("panic", rec))
			}
		}()
		err = handler(msgCtx)
	}()

	switch {
	case err == nil:
		if ackErr := d.Ack(false); ackErr != nil {
			msgLog.Error("ack failed", logger.Error(ackErr))
		}

	case errors.Is(err, ErrRequeue):
		if nackErr := d.Nack(false, true); nackErr != nil {
			msgLog.Error("nack requeue failed", logger.Error(nackErr))
		}

	default:
		if nackErr := d.Nack(false, false); nackErr != nil {
			msgLog.Error("nack drop failed", logger.Error(nackErr))
		}
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.stop != nil {
		s.stop()
	}
	conn := s.conn
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	case <-done:
	}

	if conn != nil {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("close connection: %w", err)
		}
	}

	return nil
}
