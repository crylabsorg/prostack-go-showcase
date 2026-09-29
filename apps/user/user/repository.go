package user

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
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

type pendingEvent struct {
	tx      pgx.Tx
	id      string
	subject string
	payload []byte
}

func (r *Repository) nextPending(ctx context.Context) (*pendingEvent, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	pending := &pendingEvent{tx: tx}
	err = tx.QueryRow(ctx, "SELECT id, subject, payload FROM user_outbox ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1").Scan(&pending.id, &pending.subject, &pending.payload)
	if err != nil {
		if rollbackErr := pending.rollback(ctx); rollbackErr != nil {
			return nil, errors.Join(err, rollbackErr)
		}
		return nil, err
	}
	return pending, nil
}

func (e *pendingEvent) complete(ctx context.Context) error {
	if _, err := e.tx.Exec(ctx, "DELETE FROM user_outbox WHERE id = $1", e.id); err != nil {
		return err
	}
	return e.tx.Commit(ctx)
}

func (e *pendingEvent) rollback(ctx context.Context) error {
	err := e.tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}
