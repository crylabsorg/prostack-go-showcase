package main

import (
	"context"

	"github.com/crylabsorg/prostack-go-showcase/gateway/config"
	"github.com/crylabsorg/prostack-go-showcase/gateway/user"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_grpc"
)

//	@title			ProductionStack Gateway V2 API
//	@version		2.0
//	@description	API Gateway V2 for ProductionStack microservices

//	@securityDefinitions.apikey	AccessToken
//	@in							cookie
//	@name						accessToken

//	@securityDefinitions.apikey	RefreshToken
//	@in							cookie
//	@name						refreshToken

//	@securityDefinitions.apikey	ApiKey
//	@in							header
//	@name						X-Api-Key

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

	return nil
}
