package email //nolint:testpackage // Tests exercise the private Kafka controller.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type recordingSender struct{ welcome SendWelcomeInput }

func (s *recordingSender) Send(_ context.Context, welcome SendWelcomeInput) error {
	s.welcome = welcome
	return nil
}

//nolint:paralleltest // Captures logs by replacing the process-wide zap logger.
func TestKafkaController(t *testing.T) {
	record := &kgo.Record{
		Value: []byte(`{"id":"user-id","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`),
	}
	sender := &recordingSender{}
	controller := &kafkaController{service: &service{sender: sender}}
	require.NoError(t, controller.handleMessage(context.Background(), record))
	require.Equal(
		t,
		SendWelcomeInput{ID: "user-id", FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"},
		sender.welcome,
	)
	controller.service.sender = failingSender{}
	require.ErrorContains(t, controller.handleMessage(context.Background(), record), "delivery failed")
	core, logs := observer.New(zap.ErrorLevel)
	restore := zap.ReplaceGlobals(zap.New(core))
	defer restore()
	record.Value = []byte("{")
	require.NoError(t, controller.handleMessage(context.Background(), record))
	require.Equal(t, 1, logs.Len())
}
