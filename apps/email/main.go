package main

import (
	"context"

	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_mongo"
	"github.com/crylabsorg/prostack-go/ps_nats"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
	"github.com/crylabsorg/prostack-go-showcase/email/email"
)

func main() { ps_bootstrap.Run(config.New, setup) }

func setup(ctx context.Context, drainingCtx context.Context, cfg *config.Config, app *ps_bootstrap.App) error {
	emailSender, err := email.NewSender(cfg)
	if err != nil {
		return err
	}
	emailDatabase, disconnectDatabase := ps_mongo.New(ctx, cfg.MongoURI)
	app.RegisterCleanup(func() error {
		disconnectDatabase()
		return nil
	})
	app.RegisterHealthProbe(ctx, "mongo", func(probeCtx context.Context) error {
		return emailDatabase.Client().Ping(probeCtx, nil)
	})
	emailRepository, err := email.NewRepository(ctx, emailDatabase)
	if err != nil {
		return err
	}
	if cfg.NATSURL != "" {
		natsConn, connectErr := ps_nats.Connect(cfg.NATSURL)
		if connectErr != nil {
			return connectErr
		}
		app.RegisterCleanup(func() error {
			natsConn.Close()
			return nil
		})

		app.RegisterWorker(email.RunNats(ctx, drainingCtx, natsConn, emailSender, emailRepository))
	}
	if len(cfg.KafkaBrokers) > 0 {
		app.RegisterWorker(email.RunKafka(ctx, drainingCtx, cfg.KafkaBrokers, emailSender, emailRepository))
	}

	return nil
}
