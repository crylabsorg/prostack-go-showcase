package email

import (
	"context"

	"github.com/crylabsorg/prostack-go/logger"
	"go.uber.org/zap"
)

type fakeSender struct{}

func (fakeSender) Send(ctx context.Context, w Welcome) error {
	logger.GetLogger(ctx).Info("fake welcome email", zap.String("user_id", w.ID), zap.String("to", w.Email))
	return nil
}
