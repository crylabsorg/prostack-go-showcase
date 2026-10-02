package user

import (
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"github.com/crylabsorg/prostack-go/controller"
)

func RegisterHTTP(mux *http.ServeMux, users userv1.UserServiceClient) {
	handler := httpController{Controller: controller.NewController(), users: users}
	mux.HandleFunc("POST /v1/users", handler.createUser)
	mux.HandleFunc("GET /v1/users/{id}", handler.getUser)
}

type httpController struct {
	controller.Controller
	users userv1.UserServiceClient
}

func (c httpController) createUser(w http.ResponseWriter, r *http.Request) {
	var input createUserRequest
	decoder := json.NewDecoder(r.Body)
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
	result, err := c.users.CreateUser(r.Context(), input.ToProto())
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusCreated, newUserResponse(result.GetUser()))
}

func (c httpController) getUser(w http.ResponseWriter, r *http.Request) {
	request := &userv1.GetUserRequest{}
	request.SetId(r.PathValue("id"))
	result, err := c.users.GetUser(r.Context(), request)
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusOK, newUserResponse(result.GetUser()))
}
