package email //nolint:testpackage // Verifies MongoDB initialization and private transaction cleanup.

import (
	"context"
	"errors"
	"testing"

	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRepositoryConcurrentInitialization(t *testing.T) {
	t.Parallel()
	database := ps_test.NewMongoContainer(t).DB
	const instances = 4
	results := make(chan error, instances)
	start := make(chan struct{})
	for range instances {
		go func() {
			<-start
			_, err := NewRepository(t.Context(), database)
			results <- err
		}()
	}
	close(start)
	for range instances {
		require.NoError(t, <-results)
	}
	collections, err := database.ListCollectionNames(t.Context(), bson.M{"name": "welcome_deliveries"})
	require.NoError(t, err)
	require.Equal(t, []string{"welcome_deliveries"}, collections)

	canceledCtx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = NewRepository(canceledCtx, database)
	require.ErrorIs(t, err, context.Canceled)
}

//nolint:paralleltest // Captures logs by replacing the process-wide zap logger.
func TestRepositoryCleanupUsesTransactionState(t *testing.T) {
	repository := newTestRepository(t)
	core, logs := observer.New(zap.ErrorLevel)
	restore := zap.ReplaceGlobals(zap.New(core))
	defer restore()

	t.Run("error after commit", func(t *testing.T) {
		expectedErr := errors.New("error after transaction entered committed state")
		err := repository.withinTransaction(t.Context(), func(ctx context.Context) error {
			reserved, reserveErr := repository.reserveWelcome(ctx, "committed-user")
			require.NoError(t, reserveErr)
			require.True(t, reserved)
			require.NoError(t, mongo.SessionFromContext(ctx).CommitTransaction(ctx))
			return expectedErr
		})
		require.ErrorIs(t, err, expectedErr)
		require.Zero(t, logs.Len(), "cleanup must not abort a committed transaction")
	})

	t.Run("commit error after abort", func(t *testing.T) {
		err := repository.withinTransaction(t.Context(), func(ctx context.Context) error {
			reserved, reserveErr := repository.reserveWelcome(ctx, "aborted-user")
			require.NoError(t, reserveErr)
			require.True(t, reserved)
			return mongo.SessionFromContext(ctx).AbortTransaction(ctx)
		})
		require.ErrorContains(t, err, "cannot call commitTransaction after calling abortTransaction")
		require.Zero(t, logs.Len(), "cleanup must not abort an already aborted transaction")
	})
}
