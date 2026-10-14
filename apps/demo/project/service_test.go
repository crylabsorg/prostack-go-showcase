package project_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/demo/organization"
	"github.com/crylabsorg/prostack-go-showcase/demo/project"
)

func TestProjectsAreScopedToOrganization(t *testing.T) {
	t.Parallel()
	organizations := organization.NewService()
	organizationA, err := organizations.Create("A")
	require.NoError(t, err)
	organizationB, err := organizations.Create("B")
	require.NoError(t, err)
	projects := project.NewService(organizations)

	created, err := projects.Create(organizationA.ID, " TEST ")
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	require.Equal(t, organizationA.ID, created.OrganizationID)
	require.Equal(t, "TEST", created.Name)
	fetched, err := projects.Get(organizationA.ID, created.ID)
	require.NoError(t, err)
	require.Equal(t, created, fetched)

	_, err = projects.Get(organizationB.ID, created.ID)
	require.Equal(t, project.ErrNotFound, err)
	_, err = projects.Get(organizationA.ID, "missing")
	require.Equal(t, project.ErrNotFound, err)
	_, err = projects.Create("missing", "MOCK")
	require.Equal(t, organization.ErrNotFound, err)
	_, err = projects.Create(organizationA.ID, " \t ")
	require.Equal(t, project.ErrInvalidName, err)
}

func TestConcurrentProjects(t *testing.T) {
	t.Parallel()
	organizations := organization.NewService()
	createdOrganization, err := organizations.Create("A")
	require.NoError(t, err)
	projects := project.NewService(organizations)
	for _, name := range []string{"TEST", "MOCK", "A", "B"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			created, createErr := projects.Create(createdOrganization.ID, name)
			require.NoError(t, createErr)
			fetched, getErr := projects.Get(createdOrganization.ID, created.ID)
			require.NoError(t, getErr)
			require.Equal(t, created, fetched)
		})
	}
}
