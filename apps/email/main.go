package main

import (
	"context"
	"github.com/crylabsorg/prostack-go-showcase/email/config"
	"github.com/crylabsorg/prostack-go-showcase/email/email"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
)

func main() { ps_bootstrap.Run(config.New, setup) }

func setup(ctx context.Context, cfg *config.Config, _ *ps_bootstrap.App) error {
	sender, err := email.NewSender(cfg)
	if err != nil {
		return err
	}
	go email.Run(ctx, cfg.NATSURL, sender)
	return nil
}
