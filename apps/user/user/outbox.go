package user

import (
	"context"

	"github.com/crylabsorg/prostack-go/ps_nats"
)

type Outbox struct{ service *Service }

var _ ps_nats.Outbox = (*Outbox)(nil)

func NewOutbox(service *Service) *Outbox { return &Outbox{service: service} }

func (o *Outbox) ProcessOutboxBatch(ctx context.Context, publish func([]ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult) error {
	return o.service.ProcessOutboxBatch(ctx, func(messages []OutboxMessage) []OutboxPublishResult {
		natsMessages := make([]ps_nats.OutboxMessage, len(messages))
		for i, message := range messages {
			natsMessages[i] = message
		}
		natsResults := publish(natsMessages)
		results := make([]OutboxPublishResult, len(natsResults))
		for i, result := range natsResults {
			results[i] = OutboxPublishResult{ID: result.ID(), Err: result.Err}
		}
		return results
	})
}
