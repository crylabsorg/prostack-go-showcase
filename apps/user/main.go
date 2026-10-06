package main

import (
	"context"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go-showcase/user/config"
	"github.com/crylabsorg/prostack-go-showcase/user/user"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/crylabsorg/prostack-go/ps_pg"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ps_bootstrap.Run(config.New, setup)
}

func setup(ctx context.Context, cfg *config.Config, app *ps_bootstrap.App) error {
	databasePool, disconnectDatabase := ps_pg.New(ctx, cfg.DatabaseURL)
	app.RegisterCleanup(func() error {
		disconnectDatabase()
		return nil
	})
	app.RegisterHealthProbe(ctx, "postgres", databasePool.Ping)

	userRepository := user.NewRepository(databasePool)
	userMetrics := user.NewMetrics(app.GetMetricsRegistry())
	userService := user.NewService(userRepository, userMetrics)
	natsConn, err := ps_nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	app.RegisterCleanup(func() error {
		natsConn.Close()
		return nil
	})
	jetStream, err := jetstream.New(natsConn)
	if err != nil {
		return err
	}
	if err = ps_nats.EnsureStream(ctx, jetStream, events.UserStream, "user.*"); err != nil {
		return err
	}
	go func() {
		(&ps_nats.OutboxPublisher{JS: jetStream, Outbox: userRepository}).Run(ctx)
	}()
	user.RegisterGRPC(app.GetGRPCServer(), userService)
	return nil
}
