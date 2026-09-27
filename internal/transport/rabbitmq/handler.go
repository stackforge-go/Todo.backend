package rabbitmq

// HandlerFunc — обработчик одного сообщения.
//
// Возвращаемое значение определяет, что делать:
//   - nil             → Ack
//   - ErrRequeue      → Nack(requeue=true)
//   - ErrDrop         → Nack(requeue=false), в DLQ
//   - любая другая    → Nack(requeue=false), в DLQ
type HandlerFunc func(ctx *Context) error
