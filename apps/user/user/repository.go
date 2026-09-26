package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		return fmt.Errorf("begin user creation: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)", user.ID, user.Name, user.Email)
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
	return fmt.Errorf("insert user: %w", err)
}

func (r *Repository) Get(ctx context.Context, id string) (User, error) {
	var user modelUser
	err := r.pool.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user.domain(), nil
}
