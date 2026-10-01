package user

import (
	"context"
	"net/mail"
	"strings"

	"github.com/google/uuid"
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
