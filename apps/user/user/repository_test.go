package user

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreate(t *testing.T) {
	ctx := context.Background()
	pool := ps_test.NewPostgresContainer(t, filepath.Join("..", "schema.sql"))
	repository := NewRepository(pool)
	user := User{ID: uuid.NewString(), FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"}
	require.NoError(t, repository.Create(ctx, user))
	stored, err := repository.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user, stored)
	require.Equal(t, ErrEmailExists, repository.Create(ctx, User{ID: uuid.NewString(), Email: user.Email}))
	var outboxCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM user_outbox").Scan(&outboxCount))
	require.Zero(t, outboxCount)
}

func TestRepositoryOutboxMigration(t *testing.T) {
	ctx := context.Background()
	var legacySchema []byte
	for _, filename := range []string{"20260929024205.sql", "20260929031719.sql", "20260929090000.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "migrations", filename))
		require.NoError(t, err)
		legacySchema = append(legacySchema, migration...)
	}
	schemaPath := filepath.Join(t.TempDir(), "legacy.sql")
	require.NoError(t, os.WriteFile(schemaPath, legacySchema, 0600))
	pool := ps_test.NewPostgresContainer(t, schemaPath)
	userID := uuid.NewString()
	_, err := pool.Exec(ctx, "INSERT INTO users (id, first_name, email) VALUES ($1, 'Ada', 'ada@example.com')", userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO user_outbox (id, subject, payload) VALUES ($1::uuid, $2, jsonb_build_object('id', ($1::uuid)::text, 'name', 'Ada', 'email', 'ada@example.com'))", userID, events.UserCreatedSubject)
	require.NoError(t, err)
	migration, err := os.ReadFile(filepath.Join("..", "migrations", "20261009000000.sql"))
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)

	var eventID, linkedUserID string
	var payload []byte
	var processedAt *time.Time
	require.NoError(t, pool.QueryRow(ctx, "SELECT id, user_id, payload, processed_at FROM user_outbox").Scan(&eventID, &linkedUserID, &payload, &processedAt))
	require.NotEqual(t, userID, eventID)
	require.Equal(t, userID, linkedUserID)
	require.Nil(t, processedAt)
	var event events.UserCreatedEvent
	require.NoError(t, json.Unmarshal(payload, &event))
	stored, err := NewRepository(pool).Get(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, User{ID: userID, FirstName: "Ada", Email: "ada@example.com"}, stored)
	require.Equal(t, stored.ToCreatedEvent(), event)

	// Multiple messages for the same user have independent identities.
	var nextEventID string
	require.NoError(t, pool.QueryRow(ctx, "INSERT INTO user_outbox (user_id, subject, payload) VALUES ($1, $2, $3) RETURNING id", userID, events.UserCreatedSubject, payload).Scan(&nextEventID))
	require.NotEqual(t, userID, nextEventID)
	require.NotEqual(t, eventID, nextEventID)
	_, err = pool.Exec(ctx, "INSERT INTO user_outbox (user_id, subject, payload) VALUES ($1, $2, $3)", uuid.NewString(), events.UserCreatedSubject, payload)
	require.Error(t, err)
}
