package user

import (
	"encoding/json/v2"
	"net/http"

	"github.com/crylabsorg/prostack-go/controller"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
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

// createUser creates a user.
//
//	@Summary	Create a user
//	@Tags		users
//	@Param		body	body		createUserRequest	true	"User details"
//	@Success	201		{object}	userResponse
//	@Failure	400		{object}	controller.ErrorResponse	"Invalid user input; malformed JSON returns text/plain"
//	@Failure	409		{object}	controller.ErrorResponse
//	@Failure	500		{object}	controller.ErrorResponse
//	@Router		/v1/users [post]
func (c httpController) createUser(w http.ResponseWriter, r *http.Request) {
	var input createUserRequest
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

// getUser returns a user by ID.
//
//	@Summary	Get a user
//	@Tags		users
//	@Produce	json
//	@Param		id	path		string	true	"User ID"
//	@Success	200	{object}	userResponse
//	@Failure	404	{object}	controller.ErrorResponse
//	@Failure	500	{object}	controller.ErrorResponse
//	@Router		/v1/users/{id} [get]
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
