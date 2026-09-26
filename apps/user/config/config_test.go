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
		LogPreset:           logger.PresetDevelopment,
		ServiceName:         "app1",
		HTTPListenAddr:      ":33000",
		GRPCListenAddr:      ":43000",
		DatabaseURL:         "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		NATSURL:             "nats://localhost:4222",
		OTELServiceName:     "app1",
		OTELMetricsExporter: "ps_prometheus",
		OTELTracesExporter:  "otlp",
		OTELTracesProtocol:  "grpc",
		OTELTracesEndpoint:  "http://localhost:4317",
	}

	require.Equal(t, expectedCfg, cfg)
}
