package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Service struct {
	repository *Repository
	metrics    *Metrics
}

func NewService(r *Repository, m *Metrics) *Service {
	return &Service{repository: r, metrics: m}
}

func (s *Service) Create(ctx context.Context, name, email string) (User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if name == "" || err != nil || address.Address != email {
		return User{}, ErrInvalidInput
	}
	user := User{ID: uuid.NewString(), Name: name, Email: email}
	if err := s.repository.Create(ctx, user); err != nil {
		return User{}, err
	}
	s.metrics.Created.Inc()
	return user, nil
}

func (s *Service) Get(ctx context.Context, id string) (User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return User{}, ErrInvalidInput
	}
	user, err := s.repository.Get(ctx, id)
	if err == nil {
		s.metrics.Fetched.Inc()
	}
	return user, err
}

func (s *Service) RunOutbox(ctx context.Context, js nats.JetStreamContext) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := s.publishNext(ctx, js); err != nil && !errors.Is(err, pgx.ErrNoRows) && ctx.Err() == nil {
			logger.GetLogger(ctx).Error("publish user outbox", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) publishNext(ctx context.Context, js nats.JetStreamContext) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	pending, err := s.repository.nextPending(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := pending.rollback(ctx); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("rollback user outbox: %w", rollbackErr))
		}
	}()
	if err := ps_nats.Publish(ctx, js, pending.subject, pending.id, pending.payload); err != nil {
		return fmt.Errorf("publish %s: %w", pending.id, err)
	}
	return pending.complete(ctx)
}
