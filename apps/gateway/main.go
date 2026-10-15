package main

import (
	"context"
	_ "embed"

	"github.com/crylabsorg/prostack-go/ps_asyncapi"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_grpc"
	"github.com/crylabsorg/prostack-go/ps_swagger"

	"github.com/crylabsorg/prostack-go-showcase/gateway/config"
	"github.com/crylabsorg/prostack-go-showcase/gateway/user"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

//go:embed swagger.json
var swaggerSpec []byte

//go:embed asyncapi.json
var asyncapiSpec []byte

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
	mux := app.GetHTTPServeMux()
	user.RegisterHTTP(mux, userv1.NewUserServiceClient(userConnection))

	ps_swagger.SetupSwagger(ctx, mux, swaggerSpec)
	ps_asyncapi.SetupAsyncAPI(ctx, mux, asyncapiSpec)

	return nil
}
