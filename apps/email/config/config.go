package config

import (
	"fmt"
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_config"
)

type Config struct {
	LogPreset           logger.Preset `mapstructure:"LOG_PRESET" env:"LOG_PRESET" validate:"required,oneof=production development"`
	ServiceName         string        `mapstructure:"SERVICE_NAME" env:"SERVICE_NAME" validate:"required"`
	HTTPListenAddr      string        `mapstructure:"HTTP_LISTEN_ADDR" env:"HTTP_LISTEN_ADDR" validate:"required"`
	NATSURL             string        `mapstructure:"NATS_URL" env:"NATS_URL" validate:"required"`
	Provider            string        `mapstructure:"EMAIL_PROVIDER" env:"EMAIL_PROVIDER" validate:"required,oneof=fake mailgun"`
	From                string        `mapstructure:"EMAIL_FROM" env:"EMAIL_FROM"`
	MailgunDomain       string        `mapstructure:"MAILGUN_DOMAIN" env:"MAILGUN_DOMAIN"`
	MailgunAPIKey       string        `mapstructure:"MAILGUN_API_KEY" env:"MAILGUN_API_KEY"`
	MailgunBaseURL      string        `mapstructure:"MAILGUN_BASE_URL" env:"MAILGUN_BASE_URL"`
	OTELServiceName     string        `mapstructure:"OTEL_SERVICE_NAME" env:"OTEL_SERVICE_NAME"`
	OTELMetricsExporter string        `mapstructure:"OTEL_METRICS_EXPORTER" env:"OTEL_METRICS_EXPORTER"`
	OTELTracesExporter  string        `mapstructure:"OTEL_TRACES_EXPORTER" env:"OTEL_TRACES_EXPORTER"`
	OTELTracesProtocol  string        `mapstructure:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL" env:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"`
	OTELTracesEndpoint  string        `mapstructure:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT" env:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"`
}

func (c *Config) GetLogPreset() logger.Preset { return c.LogPreset }
func (c *Config) GetServiceName() string      { return c.ServiceName }
func (c *Config) GetHTTPListenAddr() string   { return c.HTTPListenAddr }
func (c *Config) GetGRPCListenAddr() string   { return "" }

func New() (*Config, error) {
	c := &Config{}
	if err := ps_config.New(c); err != nil {
		return nil, err
	}
	if c.Provider == "mailgun" && (c.From == "" || c.MailgunDomain == "" || c.MailgunAPIKey == "") {
		return nil, fmt.Errorf("EMAIL_FROM, MAILGUN_DOMAIN and MAILGUN_API_KEY are required for Mailgun")
	}
	return c, nil
}
