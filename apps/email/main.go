package main

import (
	"context"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
	"github.com/crylabsorg/prostack-go-showcase/email/email"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_nats"
)

func main() { ps_bootstrap.Run(config.New, setup) }

func setup(ctx context.Context, cfg *config.Config, _ *ps_bootstrap.App) error {
	emailSender, err := email.NewSender(cfg)
	if err != nil {
		return err
	}
	natsClient, err := ps_nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	go func() {
		defer natsClient.Close()
		email.Run(ctx, natsClient, emailSender)
	}()
	return nil
}
