package email

import (
	"context"
	"time"
)

type Sender interface {
	Send(context.Context, Welcome) error
}

type service struct {
	sender Sender
}

func (s *service) sendWelcome(ctx context.Context, welcome Welcome) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return s.sender.Send(ctx, welcome)
}
