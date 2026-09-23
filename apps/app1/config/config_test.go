package config

import (
	"testing"

	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/testutils"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())

	cfg, err := New()
	require.NoError(t, err)
	expectedCfg := &Config{
		LogPreset: logger.PresetDevelopment,
	}

	require.Equal(t, expectedCfg, cfg)
}
