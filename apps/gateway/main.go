package main

import (
	"context"
	"log"
	"net/http"

	"github.com/crylabsorg/prostack-go-showcase/gateway/config"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
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

// @securityDefinitions.apikey	ApiKey
// @in							header
// @name						X-Api-Key

func main() {
	if err := ps_bootstrap.Run(config.New, setup); err != nil {
		log.Fatal(err)
	}
}

func setup(_ context.Context, _ *config.Config, app *ps_bootstrap.App) error {
	app.HTTP.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("gateway"))
	})
	return nil
}
