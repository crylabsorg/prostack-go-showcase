package main

import (
	"context"

	"github.com/crylabsorg/prostack-go-showcase/user/config"
	"github.com/crylabsorg/prostack-go-showcase/user/user"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_pg"
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
	go r.RunOutbox(ctx, cfg.NATSURL)
	m := user.NewMetrics(app.MetricsRegistry.Registry)
	user.RegisterGRPC(app.GRPC.Server, user.NewService(r, m))
	return nil
}
