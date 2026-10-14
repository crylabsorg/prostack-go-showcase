package main

import (
	"context"
	_ "embed"

	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_grpc"
	"github.com/crylabsorg/prostack-go/ps_swagger"

	"github.com/crylabsorg/prostack-go-showcase/gateway/config"
	"github.com/crylabsorg/prostack-go-showcase/gateway/organization"
	"github.com/crylabsorg/prostack-go-showcase/gateway/project"
	"github.com/crylabsorg/prostack-go-showcase/gateway/user"
	organizationv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/organization/v1"
	projectv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/project/v1"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

//go:embed swagger.json
var swaggerSpec []byte

func main() {
	ps_bootstrap.Run(config.New, setup)
}

func setup(ctx context.Context, _ context.Context, cfg *config.Config, app *ps_bootstrap.App) error {
	userConnection, err := ps_grpc.NewClientConn(ctx, cfg.UserGRPCAddr)
	if err != nil {
		return err
	}
	app.RegisterCleanup(func() error {
		return userConnection.Close()
	})
	demoConnection, err := ps_grpc.NewClientConn(ctx, cfg.DemoGRPCAddr)
	if err != nil {
		return err
	}
	app.RegisterCleanup(func() error {
		return demoConnection.Close()
	})

	mux := app.GetHTTPServeMux()
	user.RegisterHTTP(mux, userv1.NewUserServiceClient(userConnection))
	organization.RegisterHTTP(mux, organizationv1.NewOrganizationServiceClient(demoConnection))
	project.RegisterHTTP(mux, projectv1.NewProjectServiceClient(demoConnection))

	ps_swagger.SetupSwagger(ctx, mux, swaggerSpec)

	return nil
}
