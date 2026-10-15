package user_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/crylabsorg/prostack-go/ps_event"
	"github.com/crylabsorg/prostack-go/ps_kafka"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	usersvc "github.com/crylabsorg/prostack-go-showcase/user/user"
)

//nolint:paralleltest // Subtests exercise successive states of the same outbox and must run in order.
func TestOutboxAdapter(t *testing.T) {
	ctx := t.Context()
	pool := ps_test.NewPostgresContainer(t, filepath.Join("..", "schema.sql"))
	repository := usersvc.NewRepository(pool)
	service := usersvc.NewService(repository, usersvc.NewMetrics(prometheus.NewRegistry()))
	outbox := usersvc.NewOutbox(service)
	first, firstCreateErr := service.Create(ctx, new("Ada"), new("Lovelace"), "ada@example.com")
	require.NoError(t, firstCreateErr)
	second, secondCreateErr := service.Create(ctx, new("Grace"), new("Hopper"), "grace@example.com")
	require.NoError(t, secondCreateErr)

	var firstEventID, secondEventID string
	require.NoError(
		t,
		pool.QueryRow(ctx, "SELECT id FROM user_outbox WHERE user_id = $1", first.ID).Scan(&firstEventID),
	)
	require.NoError(
		t,
		pool.QueryRow(ctx, "SELECT id FROM user_outbox WHERE user_id = $1", second.ID).Scan(&secondEventID),
	)
	require.NotEqual(t, first.ID, firstEventID)
	require.NotEqual(t, second.ID, secondEventID)
	stored, getErr := repository.Get(ctx, first.ID)
	require.NoError(t, getErr)
	require.Equal(t, first, stored)

	t.Run("locks and partial success", func(t *testing.T) {
		publishErr := errors.New("NATS rejected message")
		err := outbox.ProcessOutboxBatch(
			ctx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 2)
				// A second worker must not publish rows already reserved by this batch.
				err := outbox.ProcessOutboxBatch(
					ctx,
					func([]ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
						t.Error("locked messages were reserved twice")
						return nil
					},
				)
				require.NoError(t, err)
				results := make([]ps_event.OutboxMessagePublishResult, 0, len(messages))
				for _, message := range messages {
					result := ps_event.OutboxMessagePublishResult{OutboxMessage: message}
					if message.ID() == secondEventID {
						result.Err = publishErr
					}
					results = append(results, result)
				}
				return results
			},
		)
		require.NoError(t, err)
		err = outbox.ProcessOutboxBatch(
			ctx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 1)
				require.Equal(t, secondEventID, messages[0].ID())
				return []ps_event.OutboxMessagePublishResult{{OutboxMessage: messages[0], Err: publishErr}}
			},
		)
		require.NoError(t, err)
		var total, processed int
		require.NoError(
			t,
			pool.QueryRow(ctx, "SELECT count(*), count(processed_at) FROM user_outbox").Scan(&total, &processed),
		)
		require.Equal(t, 2, total)
		require.Equal(t, 1, processed)
	})

	t.Run("events created during publication stay pending for the next batch", func(t *testing.T) {
		err := outbox.ProcessOutboxBatch(
			ctx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 1)
				_, err := service.Create(ctx, new("Later"), new(""), "later@example.com")
				require.NoError(t, err)
				return []ps_event.OutboxMessagePublishResult{
					{OutboxMessage: messages[0], Err: errors.New("retry later")},
				}
			},
		)
		require.NoError(t, err)
		err = outbox.ProcessOutboxBatch(
			ctx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 2)
				results := make([]ps_event.OutboxMessagePublishResult, 0, len(messages))
				for _, message := range messages {
					result := ps_event.OutboxMessagePublishResult{OutboxMessage: message}
					if message.ID() == secondEventID {
						result.Err = errors.New("retry later")
					}
					results = append(results, result)
				}
				return results
			},
		)
		require.NoError(t, err)
	})

	t.Run("cancellation releases reservation", func(t *testing.T) {
		batchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		err := outbox.ProcessOutboxBatch(
			batchCtx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 1)
				require.Equal(t, secondEventID, messages[0].ID())
				cancel()
				return []ps_event.OutboxMessagePublishResult{{OutboxMessage: messages[0]}}
			},
		)
		require.ErrorIs(t, err, context.Canceled)
		err = outbox.ProcessOutboxBatch(
			ctx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 1)
				return []ps_event.OutboxMessagePublishResult{
					{OutboxMessage: messages[0], Err: errors.New("retry later")},
				}
			},
		)
		require.NoError(t, err)
	})

	t.Run("panic releases reservation", func(t *testing.T) {
		require.Panics(t, func() {
			err := outbox.ProcessOutboxBatch(
				ctx,
				func([]ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult { panic("publish panic") },
			)
			require.NoError(t, err)
		})
		err := outbox.ProcessOutboxBatch(
			ctx,
			func(messages []ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
				require.Len(t, messages, 1)
				return []ps_event.OutboxMessagePublishResult{
					{OutboxMessage: messages[0], Err: errors.New("retry later")},
				}
			},
		)
		require.NoError(t, err)
	})

	t.Run("publishes to both brokers and retains processed events", func(t *testing.T) {
		client, err := ps_nats.Connect(ps_test.NewNATSContainer(t))
		require.NoError(t, err)
		defer client.Close()
		js, err := jetstream.New(client)
		require.NoError(t, err)
		require.NoError(t, ps_nats.EnsureStream(ctx, js, events.UserStream, "user.*"))
		consumer, err := js.CreateConsumer(
			ctx,
			events.UserStream,
			jetstream.ConsumerConfig{FilterSubject: events.UserCreatedSubject, AckPolicy: jetstream.AckExplicitPolicy},
		)
		require.NoError(t, err)
		drainingCtx, drain := context.WithCancel(ctx)
		kafkaBrokers := ps_test.NewKafkaContainer(t)
		kafkaPublisher, err := ps_kafka.NewPublisher(kafkaBrokers...)
		require.NoError(t, err)
		defer kafkaPublisher.Close()
		publisherDone := ps_event.RunOutboxPublisher(
			ctx,
			drainingCtx,
			ps_event.Fanout{ps_nats.NewPublisher(js), kafkaPublisher},
			outbox,
		)
		defer func() { drain(); <-publisherDone }()
		require.Eventually(t, func() bool {
			var count int
			countErr := pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox WHERE processed_at IS NULL").Scan(&count)
			require.NoError(t, countErr)
			return count == 0
		}, 10*time.Second, 50*time.Millisecond)
		msg, err := consumer.Next(jetstream.FetchMaxWait(3 * time.Second))
		require.NoError(t, err)
		var event events.UserCreatedEvent
		require.NoError(t, json.Unmarshal(msg.Data(), &event))
		require.Equal(t, second.ToCreatedEvent(), event)
		require.Equal(t, secondEventID, msg.Headers().Get("Nats-Msg-Id"))
		require.NoError(t, msg.DoubleAck(ctx))
		consumeCtx, cancelConsume := context.WithTimeout(ctx, 20*time.Second)
		defer cancelConsume()
		kafkaDrainCtx, drainKafka := context.WithCancel(consumeCtx)
		defer drainKafka()
		require.NoError(
			t,
			ps_kafka.RunConsumer(
				consumeCtx,
				kafkaDrainCtx,
				kafkaBrokers,
				events.UserCreatedSubject,
				"outbox-test",
				func(_ context.Context, record *kgo.Record) error {
					var kafkaEvent events.UserCreatedEvent
					require.NoError(t, json.Unmarshal(record.Value, &kafkaEvent))
					require.Equal(t, second.ToCreatedEvent(), kafkaEvent)
					require.Equal(t, secondEventID, string(record.Key))
					drainKafka()
					return nil
				},
			),
		)
		require.NoError(t, consumeCtx.Err(), "Kafka event was not delivered")
		var total, processed int
		require.NoError(
			t,
			pool.QueryRow(ctx, "SELECT count(*), count(processed_at) FROM user_outbox").Scan(&total, &processed),
		)
		require.Equal(t, 3, total)
		require.Equal(t, total, processed)
		err = outbox.ProcessOutboxBatch(ctx, func([]ps_event.OutboxMessage) []ps_event.OutboxMessagePublishResult {
			t.Error("processed messages were selected again")
			return nil
		})
		require.NoError(t, err)
	})
}
