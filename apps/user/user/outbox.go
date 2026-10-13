package user

import (
	"context"

	"github.com/crylabsorg/prostack-go/ps_event"
)

type Outbox struct{ service *Service }

var _ ps_event.Outbox = (*Outbox)(nil)

func NewOutbox(service *Service) *Outbox { return &Outbox{service: service} }

func (o *Outbox) ProcessOutboxBatch(
	ctx context.Context,
	publish func([]ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult,
) error {
	return o.service.processOutboxBatch(ctx, func(messages []OutboxMessage) []OutboxPublishResult {
		eventMessages := make([]ps_event.OutboxMessage, len(messages))
		for i, message := range messages {
			eventMessages[i] = message
		}
		eventResults := publish(eventMessages)
		results := make([]OutboxPublishResult, len(eventResults))
		for i, result := range eventResults {
			results[i] = OutboxPublishResult{ID: result.ID(), Err: result.Err}
		}
		return results
	})
}
