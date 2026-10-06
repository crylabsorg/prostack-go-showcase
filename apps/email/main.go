package main

import (
	"context"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
	"github.com/crylabsorg/prostack-go-showcase/email/email"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_nats"
)

func main() { ps_bootstrap.Run(config.New, setup) }

func setup(ctx context.Context, drainingCtx context.Context, cfg *config.Config, app *ps_bootstrap.App) error {
	emailSender, err := email.NewSender(cfg)
	if err != nil {
		return err
	}
	natsConn, err := ps_nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	app.RegisterCleanup(func() error {
		natsConn.Close()
		return nil
	})

	email.Run(ctx, drainingCtx, natsConn, emailSender)

	return nil
}
