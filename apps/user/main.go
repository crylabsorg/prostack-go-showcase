package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/crylabsorg/prostack-go-showcase/user/config"
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_grpc"
	"github.com/crylabsorg/prostack-go/ps_healthcheck"
	"github.com/crylabsorg/prostack-go/ps_http"
	"github.com/crylabsorg/prostack-go/ps_metrics"
	"github.com/crylabsorg/prostack-go/ps_shutdown"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New()
	defer logger.Init(cfg.LogPreset)()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.GetLogger(ctx).Debug("config loaded")
	logger.GetLogger(ctx).Debug("logger inialized")

	mux := http.NewServeMux()
	logger.GetLogger(ctx).Debug("mux initialized")

	grpcServer := ps_grpc.NewServer()
	logger.GetLogger(ctx).Debug("grpc server initialized")

	if err != nil {
		logger.GetLogger(ctx).Fatal("config is not loaded", zap.Error(err))
	}
	logger.GetLogger(ctx).Debug("shutdown logic prepared")

	ps_metrics.Register(mux)
	logger.GetLogger(ctx).Debug("metrics registered")
	draining := ps_shutdown.New(cancel)
	ps_healthcheck.Register(mux, ps_healthcheck.Options{
		Draining: draining,
	})
	logger.GetLogger(ctx).Debug("healthchecks registered")

	wg := sync.WaitGroup{}
	wg.Go(func() {
		if err1 := ps_http.Run(ctx, cfg.HTTPListenAddr, mux, cfg.ServiceName); err1 != nil {
			logger.GetLogger(ctx).Fatal("http server error", zap.Error(err1))
		}
	})
	wg.Go(func() {
		ps_grpc.Start(ctx, cfg.GRPCListenAddr, grpcServer)
	})
	wg.Wait()
	logger.GetLogger(ctx).Info("service gracefully stopped")
}
