package user

import (
	"context"
	"os"
	"testing"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func TestCreatePublishesUserCreated(t *testing.T) {
	databaseURL, natsURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_NATS_URL")
	if databaseURL == "" || natsURL == "" {
		t.Skip("set TEST_DATABASE_URL and TEST_NATS_URL for integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	defer pool.Close()
	r := NewRepository(pool)
	client, err := ps_nats.Connect(natsURL)
	require.NoError(t, err)
	defer client.Close()
	js, err := jetstream.New(client)
	require.NoError(t, err)
	require.NoError(t, ps_nats.EnsureStream(ctx, js, events.UserStream, "user.*"))
	consumer, err := js.CreateConsumer(ctx, events.UserStream, jetstream.ConsumerConfig{
		FilterSubject: events.UserCreatedSubject, DeliverPolicy: jetstream.DeliverNewPolicy, AckPolicy: jetstream.AckExplicitPolicy,
	})
	require.NoError(t, err)
	u := User{ID: uuid.NewString(), Name: "Ada", Email: uuid.NewString() + "@example.com"}
	defer func() {
		_, err := pool.Exec(ctx, "DELETE FROM user_outbox WHERE id = $1", u.ID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
		require.NoError(t, err)
	}()
	require.NoError(t, r.Create(ctx, u))
	var count int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox WHERE id = $1", u.ID).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, NewService(r, nil).publishNext(ctx, js))
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox WHERE id = $1", u.ID).Scan(&count))
	require.Zero(t, count)
	batch, err := consumer.Fetch(1)
	require.NoError(t, err)
	var messages []jetstream.Msg
	for msg := range batch.Messages() {
		messages = append(messages, msg)
	}
	require.NoError(t, batch.Error())
	require.Len(t, messages, 1)
	require.Contains(t, string(messages[0].Data()), u.Email)
	require.Equal(t, events.UserCreatedSubject, messages[0].Subject())
	require.NoError(t, messages[0].DoubleAck(ctx))
}
