package user

import (
	"context"
	"encoding/json/v2"
	"net/mail"
	"strings"

	"github.com/crylabsorg/prostack-go/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
)

type Service struct {
	repository *Repository
	metrics    *Metrics
}

func NewService(repository *Repository, metrics *Metrics) *Service {
	return &Service{repository: repository, metrics: metrics}
}

func (s *Service) Create(ctx context.Context, firstName, lastName *string, email string) (User, error) {
	if firstName != nil {
		trimmed := strings.TrimSpace(*firstName)
		firstName = &trimmed
	}
	if lastName != nil {
		trimmed := strings.TrimSpace(*lastName)
		lastName = &trimmed
	}
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return User{}, ErrInvalidInput
	}
	user := User{ID: uuid.NewString(), FirstName: firstName, LastName: lastName, Email: email}
	payload, err := json.Marshal(user.ToCreatedEvent())
	if err != nil {
		return User{}, err
	}
	if err = s.repository.WithinTransaction(ctx, func(repository *Repository) error {
		if createErr := repository.Create(ctx, user); createErr != nil {
			return createErr
		}
		return repository.CreateOutboxMessage(ctx, user.ID, events.UserCreatedSubject, payload)
	}); err != nil {
		return User{}, err
	}
	s.metrics.Created.Inc()
	return user, nil
}

// ProcessOutboxBatch keeps messages reserved until publication and completion.
// Only acknowledged messages are completed; failed messages remain pending.
func (s *Service) processOutboxBatch(ctx context.Context, publish func([]OutboxMessage) []OutboxPublishResult) error {
	return s.repository.WithinTransaction(ctx, func(repository *Repository) error {
		messages, err := repository.ListPendingOutboxMessages(ctx)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}
		results := publish(messages)
		acknowledgedIDs := make([]string, 0, len(results))
		for _, result := range results {
			if result.Err == nil {
				acknowledgedIDs = append(acknowledgedIDs, result.ID)
			} else {
				logger.GetLogger(ctx).
					Warn("message is not published", zap.Error(result.Err), zap.String("id", result.ID))
			}
		}
		if len(acknowledgedIDs) == 0 {
			return nil
		}
		return repository.CompleteOutboxMessages(ctx, acknowledgedIDs)
	})
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
