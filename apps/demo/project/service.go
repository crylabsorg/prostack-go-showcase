package project

import (
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/crylabsorg/prostack-go-showcase/demo/organization"
)

// Service stores demo projects in memory and keeps them scoped to an organization.
type Service struct {
	mu            sync.RWMutex
	organizations *organization.Service
	projects      map[string]Project
}

func NewService(organizations *organization.Service) *Service {
	return &Service{organizations: organizations, projects: make(map[string]Project)}
}

func (s *Service) Create(organizationID, name string) (Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, ErrInvalidName
	}
	if _, err := s.organizations.Get(organizationID); err != nil {
		return Project{}, err
	}
	project := Project{ID: uuid.NewString(), OrganizationID: organizationID, Name: name}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[project.ID] = project
	return project, nil
}

func (s *Service) Get(organizationID, id string) (Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[id]
	if !ok || project.OrganizationID != organizationID {
		return Project{}, ErrNotFound
	}
	return project, nil
}
