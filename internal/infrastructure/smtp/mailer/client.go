package mailer

import (
	"context"
	"fmt"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/smtp"
	"github.com/wneessen/go-mail"
)

type Client struct {
	client *mail.Client
	from   string
}

func New(cfg config) (*Client, error) {
	opts := []mail.Option{
		mail.WithPort(cfg.Port),
	}

	if cfg.User != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.User),
			mail.WithPassword(cfg.Password),
		)
	} else {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthNoAuth))
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("create smtp client: %w", err)
	}

	return &Client{client: client, from: cfg.From}, nil
}

func (c *Client) Send(ctx context.Context, msg smtp.Message) error {
	m := mail.NewMsg()

	if err := m.From(c.from); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := m.To(msg.To...); err != nil {
		return fmt.Errorf("set to: %w", err)
	}
	if len(msg.Cc) > 0 {
		if err := m.Cc(msg.Cc...); err != nil {
			return fmt.Errorf("set cc: %w", err)
		}
	}
	if len(msg.Bcc) > 0 {
		if err := m.Bcc(msg.Bcc...); err != nil {
			return fmt.Errorf("set bcc: %w", err)
		}
	}

	m.Subject(msg.Subject)
	m.SetBodyString(mail.TypeTextHTML, msg.Body)

	if err := c.client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
