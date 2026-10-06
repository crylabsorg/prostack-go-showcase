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
	app.Health.Probes["postgres"] = databasePool.Ping

	userRepository := user.NewRepository(databasePool)
	userMetrics := user.NewMetrics(app.MetricsRegistry.Registry)
	userService := user.NewService(userRepository, userMetrics)
	natsClient, err := ps_nats.Connect(cfg.NATSURL)
	if err != nil {
		disconnectDatabase()
		return err
	}
	jetStream, err := jetstream.New(natsClient)
	if err == nil {
		err = ps_nats.EnsureStream(ctx, jetStream, events.UserStream, "user.*")
	}
	if err != nil {
		natsClient.Close()
		disconnectDatabase()
		return err
	}
	go func() {
		defer natsClient.Close()
		defer disconnectDatabase()
		(&ps_nats.OutboxPublisher{JS: jetStream, Outbox: userRepository}).Run(ctx)
	}()
	user.RegisterGRPC(app.GRPC.Server, userService)
	return nil
}
