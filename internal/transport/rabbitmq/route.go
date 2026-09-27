package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type Route struct {
	Queue       string
	Handler     HandlerFunc
	Middleware  []Middleware
	Concurrency int
	Prefetch    int
	QueueArgs   amqp.Table
}
