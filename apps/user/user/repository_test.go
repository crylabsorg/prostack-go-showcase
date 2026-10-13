package user_test

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	usersvc "github.com/crylabsorg/prostack-go-showcase/user/user"
)

func TestRepositoryCreate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	pool := ps_test.NewPostgresContainer(t, filepath.Join("..", "schema.sql"))
	repository := usersvc.NewRepository(pool)
	user := usersvc.User{ID: uuid.NewString(), FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"}
	require.NoError(t, repository.Create(ctx, user))
	stored, err := repository.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user, stored)
	require.Equal(
		t,
		usersvc.ErrEmailExists,
		repository.Create(ctx, usersvc.User{ID: uuid.NewString(), Email: user.Email}),
	)
	var outboxCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox").Scan(&outboxCount))
	require.Zero(t, outboxCount)
}

func TestRepositoryOutboxMigration(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	migrationPaths, err := filepath.Glob(filepath.Join("..", "migrations", "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, migrationPaths)
	var schema []byte
	for _, path := range migrationPaths {
		migration, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		schema = append(schema, migration...)
		schema = append(schema, '\n')
	}
	schemaPath := filepath.Join(t.TempDir(), "migrations.sql")
	require.NoError(t, os.WriteFile(schemaPath, schema, 0600))
	pool := ps_test.NewPostgresContainer(t, schemaPath)
	repository := usersvc.NewRepository(pool)
	service := usersvc.NewService(repository, usersvc.NewMetrics(prometheus.NewRegistry()))
	created, err := service.Create(ctx, "Ada", "", "ada@example.com")
	require.NoError(t, err)
	userID := created.ID

	var eventID, linkedUserID string
	var payload []byte
	var processedAt *time.Time
	require.NoError(
		t,
		pool.QueryRow(ctx, "SELECT id, user_id, payload, processed_at FROM user_outbox").
			Scan(&eventID, &linkedUserID, &payload, &processedAt),
	)
	require.NotEqual(t, userID, eventID)
	require.Equal(t, userID, linkedUserID)
	require.Nil(t, processedAt)
	var event events.UserCreatedEvent
	require.NoError(t, json.Unmarshal(payload, &event))
	stored, err := usersvc.NewRepository(pool).Get(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, usersvc.User{ID: userID, FirstName: "Ada", Email: "ada@example.com"}, stored)
	require.Equal(t, stored.ToCreatedEvent(), event)

	// Multiple messages for the same user have independent identities.
	var nextEventID string
	require.NoError(
		t,
		pool.QueryRow(ctx, "INSERT INTO user_outbox (user_id, subject, payload) VALUES ($1, $2, $3) RETURNING id", userID, events.UserCreatedSubject, payload).
			Scan(&nextEventID),
	)
	require.NotEqual(t, userID, nextEventID)
	require.NotEqual(t, eventID, nextEventID)
	_, err = pool.Exec(
		ctx,
		"INSERT INTO user_outbox (user_id, subject, payload) VALUES ($1, $2, $3)",
		uuid.NewString(),
		events.UserCreatedSubject,
		payload,
	)
	require.Error(t, err)
}
