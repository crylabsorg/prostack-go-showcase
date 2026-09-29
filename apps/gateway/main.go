package main

import (
	"context"
	"net/http"

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

func setup(ctx context.Context, cfg *config.Config, app *ps_bootstrap.App) error {
	conn, err := ps_grpc.NewClientConn(ctx, cfg.UserGRPCAddr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	user.RegisterHTTP(app.Mux, userv1.NewUserServiceClient(conn))
	app.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("gateway"))
	})
	return nil
}
