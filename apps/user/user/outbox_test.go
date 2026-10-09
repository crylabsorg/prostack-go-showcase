package user

import (
	"context"
	"encoding/json/v2"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestOutboxAdapter(t *testing.T) {
	ctx := context.Background()
	pool := ps_test.NewPostgresContainer(t, filepath.Join("..", "schema.sql"))
	repository := NewRepository(pool)
	service := NewService(repository, NewMetrics(prometheus.NewRegistry()))
	outbox := NewOutbox(service)
	first, err := service.Create(ctx, "Ada", "Lovelace", "ada@example.com")
	require.NoError(t, err)
	second, err := service.Create(ctx, "Grace", "Hopper", "grace@example.com")
	require.NoError(t, err)

	var firstEventID, secondEventID string
	require.NoError(t, pool.QueryRow(ctx, "SELECT id FROM user_outbox WHERE user_id = $1", first.ID).Scan(&firstEventID))
	require.NoError(t, pool.QueryRow(ctx, "SELECT id FROM user_outbox WHERE user_id = $1", second.ID).Scan(&secondEventID))
	require.NotEqual(t, first.ID, firstEventID)
	require.NotEqual(t, second.ID, secondEventID)
	stored, err := repository.Get(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, first, stored)

	t.Run("locks and partial success", func(t *testing.T) {
		publishErr := errors.New("NATS rejected message")
		err := outbox.ProcessOutboxBatch(ctx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 2)
			// A second worker must not publish rows already reserved by this batch.
			err := outbox.ProcessOutboxBatch(ctx, func([]ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
				t.Error("locked messages were reserved twice")
				return nil
			})
			require.NoError(t, err)
			results := make([]ps_nats.OutboxMessagePublishResult, 0, len(messages))
			for _, message := range messages {
				result := ps_nats.OutboxMessagePublishResult{OutboxMessage: message}
				if message.ID() == secondEventID {
					result.Err = publishErr
				}
				results = append(results, result)
			}
			return results
		})
		require.NoError(t, err)
		err = outbox.ProcessOutboxBatch(ctx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 1)
			require.Equal(t, secondEventID, messages[0].ID())
			return []ps_nats.OutboxMessagePublishResult{{OutboxMessage: messages[0], Err: publishErr}}
		})
		require.NoError(t, err)
		var total, processed int
		require.NoError(t, pool.QueryRow(ctx, "SELECT count(*), count(processed_at) FROM user_outbox").Scan(&total, &processed))
		require.Equal(t, 2, total)
		require.Equal(t, 1, processed)
	})

	t.Run("events created during publication stay pending for the next batch", func(t *testing.T) {
		err := outbox.ProcessOutboxBatch(ctx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 1)
			_, err := service.Create(ctx, "Later", "", "later@example.com")
			require.NoError(t, err)
			return []ps_nats.OutboxMessagePublishResult{{OutboxMessage: messages[0], Err: errors.New("retry later")}}
		})
		require.NoError(t, err)
		err = outbox.ProcessOutboxBatch(ctx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 2)
			results := make([]ps_nats.OutboxMessagePublishResult, 0, len(messages))
			for _, message := range messages {
				result := ps_nats.OutboxMessagePublishResult{OutboxMessage: message}
				if message.ID() == secondEventID {
					result.Err = errors.New("retry later")
				}
				results = append(results, result)
			}
			return results
		})
		require.NoError(t, err)
	})

	t.Run("cancellation releases reservation", func(t *testing.T) {
		batchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		err := outbox.ProcessOutboxBatch(batchCtx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 1)
			require.Equal(t, secondEventID, messages[0].ID())
			cancel()
			return []ps_nats.OutboxMessagePublishResult{{OutboxMessage: messages[0]}}
		})
		require.ErrorIs(t, err, context.Canceled)
		err = outbox.ProcessOutboxBatch(ctx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 1)
			return []ps_nats.OutboxMessagePublishResult{{OutboxMessage: messages[0], Err: errors.New("retry later")}}
		})
		require.NoError(t, err)
	})

	t.Run("panic releases reservation", func(t *testing.T) {
		require.Panics(t, func() {
			err := outbox.ProcessOutboxBatch(ctx, func([]ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult { panic("publish panic") })
			require.NoError(t, err)
		})
		err := outbox.ProcessOutboxBatch(ctx, func(messages []ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			require.Len(t, messages, 1)
			return []ps_nats.OutboxMessagePublishResult{{OutboxMessage: messages[0], Err: errors.New("retry later")}}
		})
		require.NoError(t, err)
	})

	t.Run("publishes to JetStream and retains processed events", func(t *testing.T) {
		client, err := ps_nats.Connect(ps_test.NewNATSContainer(t))
		require.NoError(t, err)
		defer client.Close()
		js, err := jetstream.New(client)
		require.NoError(t, err)
		require.NoError(t, ps_nats.EnsureStream(ctx, js, events.UserStream, "user.*"))
		consumer, err := js.CreateConsumer(ctx, events.UserStream, jetstream.ConsumerConfig{FilterSubject: events.UserCreatedSubject, AckPolicy: jetstream.AckExplicitPolicy})
		require.NoError(t, err)
		drainingCtx, drain := context.WithCancel(ctx)
		ps_nats.RunOutboxPublisher(ctx, drainingCtx, js, outbox)
		defer drain()
		require.Eventually(t, func() bool {
			var count int
			err := pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox WHERE processed_at IS NULL").Scan(&count)
			require.NoError(t, err)
			return count == 0
		}, 10*time.Second, 50*time.Millisecond)
		msg, err := consumer.Next(jetstream.FetchMaxWait(3 * time.Second))
		require.NoError(t, err)
		var event events.UserCreatedEvent
		require.NoError(t, json.Unmarshal(msg.Data(), &event))
		require.Equal(t, second.ToCreatedEvent(), event)
		require.Equal(t, secondEventID, msg.Headers().Get("Nats-Msg-Id"))
		require.NoError(t, msg.DoubleAck(ctx))
		var total, processed int
		require.NoError(t, pool.QueryRow(ctx, "SELECT count(*), count(processed_at) FROM user_outbox").Scan(&total, &processed))
		require.Equal(t, 3, total)
		require.Equal(t, total, processed)
		err = outbox.ProcessOutboxBatch(ctx, func([]ps_nats.OutboxMessage) []ps_nats.OutboxMessagePublishResult {
			t.Error("processed messages were selected again")
			return nil
		})
		require.NoError(t, err)
	})
}
