package user

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/controller"
	"github.com/crylabsorg/prostack-go/ps_error"
	"github.com/crylabsorg/prostack-go/ps_grpc"
	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceCreate(t *testing.T) {
	traceState, err := trace.ParseTraceState("vendor=value")
	require.NoError(t, err)
	requestSpan := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled, TraceState: traceState})
	ctx := trace.ContextWithSpanContext(context.Background(), requestSpan)
	pool := ps_test.NewPostgresContainer(t, filepath.Join("..", "schema.sql"))
	repository := NewRepository(pool)
	metrics := NewMetrics(prometheus.NewRegistry())
	service := NewService(repository, metrics)

	created, err := service.Create(ctx, " Ada ", " Lovelace ", " ada@example.com ")
	require.NoError(t, err)
	stored, err := repository.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created, stored)
	require.Equal(t, "Ada", created.FirstName)
	require.Equal(t, "Lovelace", created.LastName)
	require.Equal(t, "ada@example.com", created.Email)
	var subject string
	var payload []byte
	require.NoError(t, pool.QueryRow(ctx, "SELECT subject, payload FROM user_outbox WHERE user_id = $1", created.ID).Scan(&subject, &payload))
	require.Equal(t, events.UserCreatedSubject, subject)
	var event events.UserCreatedEvent
	require.NoError(t, json.Unmarshal(payload, &event))
	require.Equal(t, created.ToCreatedEvent(), event)
	require.NoError(t, repository.WithinTransaction(context.Background(), func(repository *Repository) error {
		messages, err := repository.ListPendingOutboxMessages(context.Background())
		require.NoError(t, err)
		require.Len(t, messages, 1)
		restored := propagation.TraceContext{}.Extract(context.Background(), propagation.MapCarrier(messages[0].TraceContext()))
		require.Equal(t, requestSpan.TraceID(), trace.SpanContextFromContext(restored).TraceID())
		require.Equal(t, requestSpan.SpanID(), trace.SpanContextFromContext(restored).SpanID())
		require.Equal(t, requestSpan.TraceState().String(), trace.SpanContextFromContext(restored).TraceState().String())
		return nil
	}))
	var metric io_prometheus_client.Metric
	require.NoError(t, metrics.Created.Write(&metric))
	require.Equal(t, float64(1), metric.GetCounter().GetValue())

	_, err = service.Create(ctx, "Duplicate", "", created.Email)
	require.Equal(t, ErrEmailExists, err)

	_, err = pool.Exec(ctx, "ALTER TABLE user_outbox ADD CONSTRAINT reject_conflict CHECK (payload->>'email' <> 'conflict@example.com')")
	require.NoError(t, err)
	_, err = service.Create(ctx, "Conflict", "", "conflict@example.com")
	require.Error(t, err)
	require.NotEqual(t, ErrEmailExists, err)
	var userCount, outboxCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&userCount))
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox").Scan(&outboxCount))
	require.Equal(t, 1, userCount)
	require.Equal(t, 1, outboxCount)
	require.NoError(t, metrics.Created.Write(&metric))
	require.Equal(t, float64(1), metric.GetCounter().GetValue())
}

func TestServiceRejectsInvalidInput(t *testing.T) {
	svc := &Service{}
	for _, call := range []func() error{
		func() error {
			_, err := svc.Create(context.Background(), "Ada", "Lovelace", "not-an-email")
			return err
		},
		func() error {
			_, err := svc.Create(context.Background(), " ", "Lovelace", "ada@example.com")
			return err
		},
		func() error { _, err := svc.Get(context.Background(), "not-a-uuid"); return err },
	} {
		require.Equal(t, ErrInvalidInput, call())
	}
}

func TestUserErrorHTTPStatus(t *testing.T) {
	for _, tc := range []struct {
		err        ps_error.PSError
		httpStatus int
		grpcStatus codes.Code
	}{
		{ErrInvalidInput, http.StatusBadRequest, codes.InvalidArgument},
		{ErrNotFound, http.StatusNotFound, codes.NotFound},
		{ErrEmailExists, http.StatusConflict, codes.AlreadyExists},
	} {
		t.Run(tc.err.Message, func(t *testing.T) {
			_, rpcErr := ps_grpc.ErrorInterceptor(context.Background(), nil, nil, func(context.Context, any) (any, error) {
				return nil, fmt.Errorf("wrapped: %w", tc.err)
			})
			require.Equal(t, tc.grpcStatus, status.Code(rpcErr))
			response := httptest.NewRecorder()
			controller.NewController().WriteError(context.Background(), response, rpcErr)
			require.Equal(t, tc.httpStatus, response.Code)
		})
	}
}
