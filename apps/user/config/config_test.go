package config_test

import (
	"testing"

	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/testutils"
	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/user/config"
)

//nolint:paralleltest // t.Chdir changes the process working directory.
func TestConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())

	cfg, err := config.New()
	require.NoError(t, err)
	expectedCfg := &config.Config{
		LogPreset:      logger.PresetDevelopment,
		ServiceName:    "app1",
		HTTPListenAddr: ":33001",
		GRPCListenAddr: ":43001",
		DatabaseURL:    "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		NATSURL:        "nats://localhost:4222",
		KafkaBrokers:   []string{"localhost:9092"},
	}

	require.Equal(t, expectedCfg, cfg)
}
