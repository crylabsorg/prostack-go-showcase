package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"uuid"

	"github.com/crylabsorg/prostack-go-showcase/user/config"
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_redis"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ps_bootstrap.Run(config.New, setup)
}

func setup(ctx context.Context, _ *config.Config, app *ps_bootstrap.App) error {
	redisClient := ps_redis.New(ctx, "redis://localhost:6379")

	app.Mux.HandleFunc("GET /v1/user", func(w http.ResponseWriter, r *http.Request) {
		redisClient.Incr(r.Context(), "CALLS GET /v1/user")

		span := trace.SpanFromContext(r.Context())
		span.AddEvent("hello world!!!")
		err := errors.New("zallopa sluchilas'")
		span.RecordError(err)

		logger.GetLogger(r.Context()).Info("new request")
		type resp struct {
			ID string `json:"id"`
		}
		encoder := json.NewEncoder(w)
		w.Header().Set("Content-Type", "application/json")
		if err := encoder.Encode(resp{ID: uuid.NewV7().String()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	return nil
}
