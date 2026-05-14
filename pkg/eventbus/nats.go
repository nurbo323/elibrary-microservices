package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url,
		nats.Name("user-service"),
		nats.Timeout(5*time.Second),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &Publisher{nc: nc}, nil
}

func (p *Publisher) Publish(ctx context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	return nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		_ = p.nc.Drain()
	}
}

// UserCreatedEvent — payload subject="user.created"
type UserCreatedEvent struct {
	UserID            string    `json:"user_id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	VerificationToken string    `json:"verification_token"`
	CreatedAt         time.Time `json:"created_at"`
}

type MsgHandler func(subject string, data []byte)

func (p *Publisher) Subscribe(subject string, handler MsgHandler) error {
	_, err := p.nc.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Subject, msg.Data)
	})
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	return nil
}