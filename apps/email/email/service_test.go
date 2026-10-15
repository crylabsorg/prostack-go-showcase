package email //nolint:testpackage // Exercises private services and both transport controllers with MongoDB.

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTestRepository(t *testing.T) *Repository {
	t.Helper()
	repository, err := NewRepository(t.Context(), ps_test.NewMongoContainer(t).DB)
	require.NoError(t, err)
	return repository
}

type countingSender struct {
	calls   atomic.Int32
	err     error
	started chan struct{}
	release chan struct{}
}

func (s *countingSender) Send(ctx context.Context, _ SendWelcomeInput) error {
	s.calls.Add(1)
	if s.started != nil {
		close(s.started)
		select {
		case <-s.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.err
}

type unacknowledgedMessage struct{ acknowledgedMessage }

func (*unacknowledgedMessage) DoubleAck(context.Context) error {
	return errors.New("acknowledgement failed")
}

func TestWelcomeDeliveryDeduplicatesAcrossBrokersAndRestarts(t *testing.T) {
	t.Parallel()
	repository := newTestRepository(t)
	sender := &countingSender{}
	ctx := t.Context()
	payload := []byte(`{"id":"user-id","email":"ada@example.com"}`)
	natsController := &natsController{service: &service{sender: sender, repository: repository}}
	failedAck := &unacknowledgedMessage{data: payload}
	require.ErrorContains(t, natsController.handleMessage(ctx, failedAck), "acknowledgement failed")
	// Redelivery after a failed acknowledgement must not resend the committed email.
	message := &acknowledgedMessage{data: payload}
	require.NoError(t, natsController.handleMessage(ctx, message))
	require.True(t, message.acknowledged)
	message.acknowledged = false
	require.NoError(t, natsController.handleMessage(ctx, message))
	require.True(t, message.acknowledged)

	// A fresh repository and service see the same persisted history.
	restartedRepository, err := NewRepository(ctx, repository.deliveries.Database())
	require.NoError(t, err)
	kafkaController := &kafkaController{service: &service{
		sender:     sender,
		repository: restartedRepository,
	}}
	record := &kgo.Record{Value: payload}
	require.NoError(t, kafkaController.handleMessage(ctx, record))
	require.NoError(t, kafkaController.handleMessage(ctx, record))
	require.EqualValues(t, 1, sender.calls.Load())
	var stored modelWelcomeDelivery
	require.NoError(t, repository.deliveries.FindOne(ctx, bson.M{"_id": "user-id"}).Decode(&stored))
	require.NotNil(t, stored.SentAt)

	// A different user still receives a welcome email.
	record.Value = []byte(`{"id":"another-user","email":"grace@example.com"}`)
	require.NoError(t, kafkaController.handleMessage(ctx, record))
	require.EqualValues(t, 2, sender.calls.Load())
}

func TestWelcomeDeliveryRetriesAfterFailure(t *testing.T) {
	t.Parallel()
	repository := newTestRepository(t)
	sender := &countingSender{err: errors.New("provider unavailable")}
	service := &service{sender: sender, repository: repository}
	welcome := SendWelcomeInput{ID: "user-id", Email: "ada@example.com"}
	require.ErrorContains(t, service.sendWelcome(t.Context(), welcome), "provider unavailable")
	count, err := repository.deliveries.CountDocuments(t.Context(), bson.M{})
	require.NoError(t, err)
	require.Zero(t, count)
	sender.err = nil
	require.NoError(t, service.sendWelcome(t.Context(), welcome))
	require.NoError(t, service.sendWelcome(t.Context(), welcome))
	require.EqualValues(t, 2, sender.calls.Load())
}

func TestWelcomeDeliveryConcurrentConsumers(t *testing.T) {
	t.Parallel()
	repository := newTestRepository(t)
	sender := &countingSender{started: make(chan struct{}), release: make(chan struct{})}
	firstService := &service{sender: sender, repository: repository}
	secondRepository, err := NewRepository(t.Context(), repository.deliveries.Database())
	require.NoError(t, err)
	secondService := &service{sender: sender, repository: secondRepository}
	welcome := SendWelcomeInput{ID: "user-id", Email: "ada@example.com"}
	done := make(chan error, 1)
	go func() { done <- firstService.sendWelcome(t.Context(), welcome) }()
	select {
	case <-sender.started:
	case <-time.After(5 * time.Second):
		require.FailNow(t, "first consumer did not start delivery")
	}
	// The competing transaction must fail before reaching the provider.
	require.Error(t, secondService.sendWelcome(t.Context(), welcome))
	require.EqualValues(t, 1, sender.calls.Load())
	close(sender.release)
	require.NoError(t, <-done)
	require.NoError(t, secondService.sendWelcome(t.Context(), welcome))
	require.EqualValues(t, 1, sender.calls.Load())
}

func TestWelcomeDeliveryCanceledAttemptCanRetry(t *testing.T) {
	t.Parallel()
	repository := newTestRepository(t)
	sender := &countingSender{started: make(chan struct{}), release: make(chan struct{})}
	service := &service{sender: sender, repository: repository}
	welcome := SendWelcomeInput{ID: "user-id", Email: "ada@example.com"}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- service.sendWelcome(ctx, welcome) }()
	select {
	case <-sender.started:
	case <-time.After(5 * time.Second):
		require.FailNow(t, "delivery did not start")
	}
	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	service.sender = &countingSender{}
	require.NoError(t, service.sendWelcome(t.Context(), welcome))
}
