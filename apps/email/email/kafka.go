package email

import (
	"context"
	"encoding/json/v2"
	"time"

	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
)

type kafkaController struct{ service *service }

func RunKafka(ctx, drainingCtx context.Context, brokers []string, sender Sender) {
	controller := &kafkaController{service: &service{sender: sender}}
	go func() {
		for ctx.Err() == nil && drainingCtx.Err() == nil {
			err := ps_kafka.RunConsumer(
				ctx,
				drainingCtx,
				brokers,
				events.UserCreatedSubject,
				"email-sender",
				controller.handleMessage,
			)
			if err != nil {
				logger.GetLogger(ctx).Error("email kafka consumer", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-drainingCtx.Done():
				return
			case <-time.After(consumerRetryDelay):
			}
		}
	}()
}

func (c *kafkaController) handleMessage(ctx context.Context, record *kgo.Record) error {
	var event events.UserCreatedEvent
	if err := json.Unmarshal(record.Value, &event); err != nil {
		// Commit malformed events, matching Term() in the NATS controller.
		logger.GetLogger(ctx).
			Error("invalid user.created kafka event", zap.Int32("partition", record.Partition), zap.Int64("offset", record.Offset), zap.Error(err))
		return nil
	}
	return c.service.sendWelcome(
		ctx,
		SendWelcomeInput{ID: event.ID, FirstName: event.FirstName, LastName: event.LastName, Email: event.Email},
	)
}
