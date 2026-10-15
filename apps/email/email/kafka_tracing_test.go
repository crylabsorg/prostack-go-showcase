package email //nolint:testpackage // Tests exercise the private Kafka controller.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestKafkaMalformedEventTracing(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() {
		require.NoError(t, provider.Shutdown(context.Background()))
	})
	ctx, span := provider.Tracer("test").Start(
		context.Background(),
		"user.created process",
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	controller := &kafkaController{}

	// Malformed events are committed without calling the service, but retain the failure in the span.
	require.NoError(t, controller.handleMessage(ctx, &kgo.Record{Value: []byte("{")}))
	span.End()

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	require.Equal(t, codes.Error, ended[0].Status().Code)
	require.Equal(t, "invalid user.created kafka event", ended[0].Status().Description)
	require.Contains(t, ended[0].Attributes(), attribute.String("error.type", "*jsontext.SyntacticError"))
	events := ended[0].Events()
	require.Len(t, events, 1)
	require.Equal(t, "exception", events[0].Name)
	require.Contains(t, events[0].Attributes, attribute.String("exception.type", "*jsontext.SyntacticError"))
	var exceptionMessage string
	for _, attr := range events[0].Attributes {
		if attr.Key == "exception.message" {
			exceptionMessage = attr.Value.AsString()
		}
	}
	require.Contains(t, exceptionMessage, "unexpected EOF")
}
