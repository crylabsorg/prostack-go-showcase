package config

import (
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_config"
)

type Config struct {
	LogPreset      logger.Preset `mapstructure:"LOG_PRESET" env:"LOG_PRESET" validate:"required,oneof=production development"`
	ServiceName    string        `mapstructure:"SERVICE_NAME" env:"SERVICE_NAME" validate:"required"`
	HTTPListenAddr string        `mapstructure:"HTTP_LISTEN_ADDR" env:"HTTP_LISTEN_ADDR" validate:"required"`
	GRPCListenAddr string        `mapstructure:"GRPC_LISTEN_ADDR" env:"GRPC_LISTEN_ADDR" validate:"required"`
	DatabaseURL    string        `mapstructure:"DATABASE_URL" env:"DATABASE_URL" validate:"required"`

	OTEL_ServiceName                string `mapstructure:"OTEL_SERVICE_NAME" env:"OTEL_SERVICE_NAME"`
	OTEL_MetricsExporter            string `mapstructure:"OTEL_METRICS_EXPORTER" env:"OTEL_METRICS_EXPORTER"`
	OTEL_TracesExporter             string `mapstructure:"OTEL_TRACES_EXPORTER" env:"OTEL_TRACES_EXPORTER"`
	OTEL_ExporterOTLPTracesProtocol string `mapstructure:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL" env:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"`
	OTEL_ExporterOTLPTracesEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT" env:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"`
}

func (c *Config) GetLogPreset() logger.Preset { return c.LogPreset }
func (c *Config) GetServiceName() string      { return c.ServiceName }
func (c *Config) GetHTTPListenAddr() string   { return c.HTTPListenAddr }
func (c *Config) GetGRPCListenAddr() string   { return c.GRPCListenAddr }

func New() (*Config, error) {
	cfg := &Config{}
	err := ps_config.New(cfg)
	return cfg, err
}
