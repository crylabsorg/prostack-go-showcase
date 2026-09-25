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
		LogPreset:                       logger.PresetDevelopment,
		ServiceName:                     "app1",
		HTTPListenAddr:                  ":33000",
		GRPCListenAddr:                  ":43000",
		OTEL_ServiceName:                "app1",
		OTEL_MetricsExporter:            "ps_prometheus",
		OTEL_TracesExporter:             "otlp",
		OTEL_ExporterOTLPTracesProtocol: "grpc",
		OTEL_ExporterOTLPTracesEndpoint: "http://localhost:4317",
	}

	require.Equal(t, expectedCfg, cfg)
}
