package user

import (
	"context"
	"errors"
	"time"

	"github.com/crylabsorg/prostack-go/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// database is implemented by both the connection pool and a transaction.
type database interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Repository struct{ db database }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{db: pool} }

func (r *Repository) Create(ctx context.Context, user User) error {
	_, err := r.db.Exec(ctx, "INSERT INTO users (id, first_name, last_name, email) VALUES ($1, $2, $3, $4)", user.ID, user.FirstName, user.LastName, user.Email)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
		return ErrEmailExists
	}
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (User, error) {
	var user modelUser
	err := r.db.QueryRow(ctx, "SELECT id, COALESCE(first_name, ''), COALESCE(last_name, ''), email FROM users WHERE id = $1", id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user.toDomain(), nil
}

// WithinTransaction binds the repository to one transaction for fn.
// A callback error or panic rolls back all writes.
func (r *Repository) WithinTransaction(ctx context.Context, fn func(*Repository) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackTransaction(ctx, tx)
	if err := fn(&Repository{db: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) AppendOutbox(ctx context.Context, userID, subject string, payload []byte) error {
	_, err := r.db.Exec(ctx, "INSERT INTO user_outbox (user_id, subject, payload) VALUES ($1, $2, $3)", userID, subject, payload)
	return err
}

// PendingOutbox must run within WithinTransaction to keep rows locked until completion.
func (r *Repository) PendingOutbox(ctx context.Context) ([]OutboxMessage, error) {
	const limit = 100
	rows, err := r.db.Query(ctx, "SELECT id, subject, payload FROM user_outbox WHERE processed_at IS NULL ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]OutboxMessage, 0, limit)
	for rows.Next() {
		var message OutboxMessage
		if err := rows.Scan(&message.id, &message.subject, &message.payload); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *Repository) CompleteOutbox(ctx context.Context, ids []string) error {
	_, err := r.db.Exec(ctx, "UPDATE user_outbox SET processed_at = now() WHERE id = ANY($1::uuid[])", ids)
	return err
}

func rollbackTransaction(ctx context.Context, tx pgx.Tx) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := tx.Rollback(cleanupCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		logger.GetLogger(ctx).Error("rollback user transaction", zap.Error(err))
	}
}
