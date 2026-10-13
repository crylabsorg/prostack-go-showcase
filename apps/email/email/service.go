package email

import (
	"context"
	"time"
)

const deliveryTimeout = 20 * time.Second

type Sender interface {
	Send(context.Context, SendWelcomeInput) error
}

type service struct {
	sender Sender
}

func (s *service) sendWelcome(ctx context.Context, welcome SendWelcomeInput) error {
	ctx, cancel := context.WithTimeout(ctx, deliveryTimeout)
	defer cancel()
	return s.sender.Send(ctx, welcome)
}
