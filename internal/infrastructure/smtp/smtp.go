package smtp

import "context"

type Message struct {
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
}

type Client interface {
	Send(ctx context.Context, msg Message) error
}
