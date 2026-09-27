package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

type Context struct {
	ctx      context.Context
	delivery amqp.Delivery
	ch       *amqp.Channel
	logger   logger.Logger
}

func newContext(
	ctx context.Context,
	d amqp.Delivery,
	ch *amqp.Channel,
	log logger.Logger,
) *Context {
	return &Context{
		ctx:      ctx,
		delivery: d,
		ch:       ch,
		logger:   log,
	}
}

func (c *Context) Context() context.Context { return c.ctx }
func (c *Context) Delivery() amqp.Delivery  { return c.delivery }
func (c *Context) Body() []byte             { return c.delivery.Body }
func (c *Context) RoutingKey() string       { return c.delivery.RoutingKey }
func (c *Context) MessageID() string        { return c.delivery.MessageId }
func (c *Context) Logger() logger.Logger    { return c.logger }
func (c *Context) Channel() *amqp.Channel   { return c.ch }

func (c *Context) BindJSON(dest any) error {
	return json.Unmarshal(c.delivery.Body, dest)
}

func (c *Context) Requeue(err error) error {
	if err == nil {
		return ErrRequeue
	}
	return err
}

func (c *Context) Drop(err error) error {
	if err == nil {
		return ErrDrop
	}
	return err
}
