package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/crylabsorg/prostack-go-showcase/gateway/config"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"github.com/crylabsorg/prostack-go/ps_bootstrap"
	"github.com/crylabsorg/prostack-go/ps_grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	users := userv1.NewUserServiceClient(conn)
	app.Mux.HandleFunc("POST /v1/users", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		user, err := users.CreateUser(r.Context(), &userv1.CreateUserRequest{Name: input.Name, Email: input.Email})
		if err != nil {
			writeRPCError(w, err)
			return
		}
		writeUser(w, http.StatusCreated, user.GetUser())
	})
	app.Mux.HandleFunc("GET /v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, err := users.GetUser(r.Context(), &userv1.GetUserRequest{Id: r.PathValue("id")})
		if err != nil {
			writeRPCError(w, err)
			return
		}
		writeUser(w, http.StatusOK, user.GetUser())
	})
	app.Mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("gateway"))
	})
	return nil
}

func writeUser(w http.ResponseWriter, code int, user *userv1.User) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{user.GetId(), user.GetName(), user.GetEmail()})
}

func writeRPCError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		http.Error(w, status.Convert(err).Message(), http.StatusBadRequest)
	case codes.NotFound:
		http.Error(w, "user not found", http.StatusNotFound)
	case codes.AlreadyExists:
		http.Error(w, "email already exists", http.StatusConflict)
	default:
		http.Error(w, "user service unavailable", http.StatusBadGateway)
	}
}
