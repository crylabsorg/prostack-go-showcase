package email

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type failingSender struct{}

func (failingSender) Send(context.Context, SendWelcomeInput) error {
	return errors.New("delivery failed")
}

type failingMessage struct {
	jetstream.Msg
	data []byte
}

func (m failingMessage) Data() []byte         { return m.data }
func (m failingMessage) Headers() nats.Header { return nil }
func (m failingMessage) Subject() string      { return "user.created" }
func (m failingMessage) Term() error          { return errors.New("terminate message") }
func (m failingMessage) Nak() error           { return errors.New("nak message") }

func TestNATSControllerReportsAcknowledgementFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		data   string
		sender Sender
		want   string
		logged bool
	}{
		{name: "term", data: "{", want: "terminate message"},
		{name: "nak", data: `{"id":"1","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`, sender: failingSender{}, want: "nak message", logged: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			core, logs := observer.New(zap.ErrorLevel)
			restore := zap.ReplaceGlobals(zap.New(core))
			defer restore()

			msg := failingMessage{data: []byte(test.data)}
			err := (&natsController{service: &service{sender: test.sender}}).handleMessage(context.Background(), msg)
			require.ErrorContains(t, err, test.want)

			if test.logged {
				require.Equal(t, 1, logs.Len())
				require.Contains(t, logs.All()[0].ContextMap()["error"], test.want)
			}
		})
	}
}
