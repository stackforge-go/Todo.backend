package rabbitmq

import "errors"

var (
	ErrRequeue = errors.New("requeue")
	ErrDrop    = errors.New("drop")
)
