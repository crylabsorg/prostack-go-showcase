package config

import (
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_config"
)

type Config struct {
	LogPreset      logger.Preset `mapstructure:"LOG_PRESET"       env:"LOG_PRESET"       validate:"required,oneof=production development"`
	ServiceName    string        `mapstructure:"SERVICE_NAME"     env:"SERVICE_NAME"     validate:"required"`
	HTTPListenAddr string        `mapstructure:"HTTP_LISTEN_ADDR" env:"HTTP_LISTEN_ADDR" validate:"required"`
	GRPCListenAddr string        `mapstructure:"GRPC_LISTEN_ADDR" env:"GRPC_LISTEN_ADDR"`
	UserGRPCAddr   string        `mapstructure:"USER_GRPC_ADDR"   env:"USER_GRPC_ADDR"   validate:"required"`
	DemoGRPCAddr   string        `mapstructure:"DEMO_GRPC_ADDR"   env:"DEMO_GRPC_ADDR"   validate:"required"`
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
