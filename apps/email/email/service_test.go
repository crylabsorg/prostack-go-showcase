package email

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type failingSender struct{}

func (failingSender) Send(context.Context, Welcome) error { return errors.New("delivery failed") }

func TestServiceLogsAcknowledgementFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		data   string
		sender Sender
		want   string
	}{
		{name: "term", data: "{", want: "terminate message"},
		{name: "nak", data: `{"ID":"1","Name":"Ada","Email":"ada@example.com"}`, sender: failingSender{}, want: "nak message"},
	} {
		t.Run(test.name, func(t *testing.T) {
			core, logs := observer.New(zap.ErrorLevel)
			restore := zap.ReplaceGlobals(zap.New(core))
			defer restore()

			msg := nats.NewMsg("user.created")
			msg.Data = []byte(test.data)
			(&service{sender: test.sender}).handleMessage(context.Background(), msg)

			require.Equal(t, 1, logs.Len())
			require.Contains(t, logs.All()[0].ContextMap()["error"], test.want)
		})
	}
}
