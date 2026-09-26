package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// RunOutbox retries until both NATS and pending messages are available.
func (r *Repository) RunOutbox(ctx context.Context, url string) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var conn *nats.Conn
	var js nats.JetStreamContext
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	for {
		if conn == nil || !conn.IsConnected() {
			if conn != nil {
				conn.Close()
			}
			var err error
			conn, err = ps_nats.Connect(url)
			if err != nil {
				logger.GetLogger(ctx).Warn("connect to NATS", zap.Error(err))
			} else if js, err = conn.JetStream(); err == nil {
				err = ps_nats.EnsureStream(js, events.UserStream, "user.*")
			}
			if err != nil && conn != nil {
				logger.GetLogger(ctx).Warn("create user event stream", zap.Error(err))
				conn.Close()
				conn = nil
			}
		} else if err := r.publishNext(ctx, js); err != nil && !errors.Is(err, pgx.ErrNoRows) && ctx.Err() == nil {
			logger.GetLogger(ctx).Error("publish user outbox", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Repository) publishNext(ctx context.Context, js nats.JetStreamContext) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id string
	var subject string
	var payload []byte
	err = tx.QueryRow(ctx, "SELECT id, subject, payload FROM user_outbox ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1").Scan(&id, &subject, &payload)
	if err != nil {
		return err
	}
	if err = ps_nats.Publish(ctx, js, subject, id, payload); err != nil {
		return fmt.Errorf("publish %s: %w", id, err)
	}
	if _, err = tx.Exec(ctx, "DELETE FROM user_outbox WHERE id = $1", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
