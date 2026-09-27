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
	pool, disconnect := ps_pg.New(ctx, cfg.DatabaseURL)
	app.Health.Probes["postgres"] = pool.Ping
	go func() {
		<-ctx.Done()
		disconnect()
	}()

	r := user.NewRepository(pool)
	m := user.NewMetrics(app.MetricsRegistry.Registry)
	s := user.NewService(r, m)
	client, err := ps_nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	js, err := jetstream.New(client)
	if err == nil {
		err = ps_nats.EnsureStream(ctx, js, events.UserStream, "user.*")
	}
	if err != nil {
		client.Close()
		return err
	}
	go func() {
		defer client.Close()
		s.RunOutbox(ctx, js)
	}()
	user.RegisterGRPC(app.GRPC.Server, s)
	return nil
}
