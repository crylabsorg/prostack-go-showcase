package user

import (
	"context"
	"encoding/json/v2"
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
	return user.toDomain(), nil
}

type pendingEvent struct {
	id      string
	subject string
	payload []byte
	tx      pgx.Tx
}

func (e *pendingEvent) ID() string {
	return e.id
}

func (e *pendingEvent) Subject() string {
	return e.subject
}

func (e *pendingEvent) Payload() []byte {
	return e.payload
}

func (r *Repository) NextBatch(ctx context.Context) ([]pendingEvent, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	const limit = 100
	rows, err := tx.Query(ctx, "SELECT id, subject, payload FROM user_outbox ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED", limit)
	if err != nil {
		return nil, err
	}
	events := make([]pendingEvent, 0, limit)
	for rows.Next() {
		var event pendingEvent
		if err := rows.Scan(&event.id, &event.subject, &event.payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *Repository) Complete(ctx context.Context, success []ps_nats.OutboxMessage, failed []ps_nats.OutboxMessage, failedReasons []error) error {

	return nil
}
