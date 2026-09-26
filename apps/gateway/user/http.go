package user

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"github.com/crylabsorg/prostack-go/controller"
)

func RegisterHTTP(mux *http.ServeMux, users userv1.UserServiceClient) {
	responses := controller.NewController()
	mux.HandleFunc("POST /v1/users", func(w http.ResponseWriter, r *http.Request) {
		var input createUserRequest
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
		result, err := users.CreateUser(r.Context(), input.ToProto())
		if err != nil {
			responses.WriteError(r.Context(), w, err)
			return
		}
		writeUser(w, http.StatusCreated, result.GetUser())
	})
	mux.HandleFunc("GET /v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		request := &userv1.GetUserRequest{}
		request.SetId(r.PathValue("id"))
		result, err := users.GetUser(r.Context(), request)
		if err != nil {
			responses.WriteError(r.Context(), w, err)
			return
		}
		writeUser(w, http.StatusOK, result.GetUser())
	})
}

func writeUser(w http.ResponseWriter, code int, user *userv1.User) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(userResponseFromProto(user))
}
