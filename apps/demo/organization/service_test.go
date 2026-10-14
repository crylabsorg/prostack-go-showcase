package organization_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/demo/organization"
)

func TestOrganizations(t *testing.T) {
	t.Parallel()
	service := organization.NewService()
	created, err := service.Create(" A ")
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	require.Equal(t, "A", created.Name)

	fetched, err := service.Get(created.ID)
	require.NoError(t, err)
	require.Equal(t, created, fetched)

	_, err = service.Create(" \t ")
	require.Equal(t, organization.ErrInvalidName, err)
	_, err = service.Get("missing")
	require.Equal(t, organization.ErrNotFound, err)

	// Separate application instances have separate, ephemeral data.
	_, err = organization.NewService().Get(created.ID)
	require.Equal(t, organization.ErrNotFound, err)
}
