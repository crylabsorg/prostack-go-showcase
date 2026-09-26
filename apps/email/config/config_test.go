package config

import (
	"testing"

	"github.com/crylabsorg/prostack-go/testutils"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())
	c, err := New()
	require.NoError(t, err)
	require.Equal(t, "fake", c.Provider)
	require.Equal(t, ":33002", c.HTTPListenAddr)
}
