package user

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Create(ctx context.Context, user User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "INSERT INTO users (id, first_name, email) VALUES ($1, $2, $3)", user.ID, user.Name, user.Email)
	if err == nil {
		payload, marshalErr := json.Marshal(user.ToCreatedEvent())
		if marshalErr != nil {
			return marshalErr
		}
		_, err = tx.Exec(ctx, "INSERT INTO user_outbox (id, subject, payload) VALUES ($1, $2, $3)", user.ID, events.UserCreatedSubject, payload)
		if err == nil {
			return tx.Commit(ctx)
		}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrEmailExists
	}
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (User, error) {
	var user modelUser
	err := r.pool.QueryRow(ctx, "SELECT id, first_name, email FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user.domain(), nil
}

type pendingEvents struct {
	tx       pgx.Tx
	events   []pendingEvent
	finished bool
}

type pendingEvent struct {
	id      string
	subject string
	payload []byte
}

func (r *Repository) NextBatch(ctx context.Context, limit int) ([]ps_nats.OutboxMessage, error) {
	if limit < 1 {
		return nil, errors.New("outbox batch size must be positive")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, "SELECT id, subject, payload FROM user_outbox ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED", limit)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return nil, errors.Join(err, rollbackErr)
		}
		return nil, err
	}
	defer rows.Close()
	batch := &pendingEvents{tx: tx}
	for rows.Next() {
		var event pendingEvent
		if err := rows.Scan(&event.id, &event.subject, &event.payload); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				return nil, errors.Join(err, rollbackErr)
			}
			return nil, err
		}
		batch.events = append(batch.events, event)
	}
	if err := rows.Err(); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return nil, errors.Join(err, rollbackErr)
		}
		return nil, err
	}
	if len(batch.events) == 0 {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			return nil, err
		}
		return nil, ps_nats.ErrOutboxEmpty
	}
	messages := make([]ps_nats.OutboxMessage, len(batch.events))
	for i := range batch.events {
		messages[i] = &pendingBatchEvent{batch: batch, index: i}
	}
	return messages, nil
}

type pendingBatchEvent struct {
	batch *pendingEvents
	index int
}

func (e *pendingBatchEvent) ID() string      { return e.batch.events[e.index].id }
func (e *pendingBatchEvent) Subject() string { return e.batch.events[e.index].subject }
func (e *pendingBatchEvent) Payload() []byte { return e.batch.events[e.index].payload }

func (r *Repository) CompleteBatch(ctx context.Context, messages []ps_nats.OutboxMessage) error {
	b := messages[0].(*pendingBatchEvent).batch
	if b.finished {
		return nil
	}
	ids := make([]string, len(messages))
	for i, message := range messages {
		ids[i] = message.ID()
	}
	if _, err := b.tx.Exec(ctx, "DELETE FROM user_outbox WHERE id = ANY($1::uuid[])", ids); err != nil {
		return err
	}
	if err := b.tx.Commit(ctx); err != nil {
		return err
	}
	b.finished = true
	return nil
}

func (r *Repository) RollbackBatch(ctx context.Context, messages []ps_nats.OutboxMessage) error {
	b := messages[0].(*pendingBatchEvent).batch
	if b.finished {
		return nil
	}
	b.finished = true
	err := b.tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}
