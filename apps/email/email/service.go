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
	sender     Sender
	repository *Repository
}

func (s *service) sendWelcome(ctx context.Context, welcome SendWelcomeInput) error {
	ctx, cancel := context.WithTimeout(ctx, deliveryTimeout)
	defer cancel()
	return s.repository.withinTransaction(ctx, func(transactionCtx context.Context) error {
		reserved, err := s.repository.reserveWelcome(transactionCtx, welcome.ID)
		if err != nil || !reserved {
			return err
		}
		if err = s.sender.Send(transactionCtx, welcome); err != nil {
			return err
		}
		return s.repository.completeWelcome(transactionCtx, welcome.ID)
	})
}
