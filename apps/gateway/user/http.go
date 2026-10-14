package user

import (
	"encoding/json/v2"
	"net/http"

	"github.com/crylabsorg/prostack-go/controller"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

func RegisterHTTP(mux *http.ServeMux, users userv1.UserServiceClient) {
	handler := httpController{Controller: controller.NewController(), users: users}
	HandlerFromMux(handler, mux)
}

type httpController struct {
	controller.Controller

	users userv1.UserServiceClient
}

func (c httpController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input CreateUserRequest
	if err := json.UnmarshalRead(r.Body, &input, json.RejectUnknownMembers(true)); err != nil {
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

func (c httpController) GetUser(w http.ResponseWriter, r *http.Request, id string) {
	request := &userv1.GetUserRequest{}
	request.SetId(id)
	result, err := c.users.GetUser(r.Context(), request)
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusOK, newUserResponse(result.GetUser()))
}
