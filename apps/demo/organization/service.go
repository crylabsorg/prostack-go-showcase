package organization

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

// Service stores demo organizations in memory until the application restarts.
type Service struct {
	mu            sync.RWMutex
	organizations map[string]Organization
}

func NewService() *Service {
	return &Service{organizations: make(map[string]Organization)}
}

func (s *Service) Create(name string) (Organization, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Organization{}, ErrInvalidName
	}
	organization := Organization{ID: uuid.NewString(), Name: name}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.organizations[organization.ID] = organization
	return organization, nil
}

func (s *Service) Get(id string) (Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	organization, ok := s.organizations[id]
	if !ok {
		return Organization{}, ErrNotFound
	}
	return organization, nil
}
