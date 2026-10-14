package main

import (
	"context"

	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"google.golang.org/grpc"

	"github.com/crylabsorg/prostack-go-showcase/demo/config"
	"github.com/crylabsorg/prostack-go-showcase/demo/organization"
	"github.com/crylabsorg/prostack-go-showcase/demo/project"
)

func main() {
	ps_bootstrap.Run(config.New, setup)
}

func setup(_ context.Context, _ context.Context, _ *config.Config, app *ps_bootstrap.App) error {
	organizationService := organization.NewService()
	projectService := project.NewService(organizationService)
	app.RegisterGRPC(func(grpcServer *grpc.Server) {
		organization.RegisterGRPC(grpcServer, organizationService)
		project.RegisterGRPC(grpcServer, projectService)
	})
	return nil
}
