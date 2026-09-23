package config

import (
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_config"
)

type Config struct {
	LogPreset      logger.Preset `mapstructure:"LOG_PRESET"       env:"LOG_PRESET"       validate:"required,oneof=production development"`
	ServiceName    string        `mapstructure:"SERVICE_NAME" env:"SERVICE_NAME" validate:"required"`
	HTTPListenAddr string        `mapstructure:"HTTP_LISTEN_ADDR" env:"HTTP_LISTEN_ADDR" validate:"required"`
}

func New() (*Config, error) {
	cfg := &Config{}
	err := ps_config.New(cfg)
	return cfg, err
}
