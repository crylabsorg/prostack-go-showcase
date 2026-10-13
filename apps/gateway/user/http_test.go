package user_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/crylabsorg/prostack-go-showcase/gateway/user"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

type testUserClient struct {
	userv1.UserServiceClient

	createRequest *userv1.CreateUserRequest
	getRequest    *userv1.GetUserRequest
}

func (client *testUserClient) CreateUser(
	_ context.Context,
	request *userv1.CreateUserRequest,
	_ ...grpc.CallOption,
) (*userv1.CreateUserResponse, error) {
	client.createRequest = request
	user := &userv1.User{}
	user.SetId("user-id")
	user.SetFirstName(request.GetFirstName())
	user.SetLastName(request.GetLastName())
	user.SetEmail(request.GetEmail())
	response := &userv1.CreateUserResponse{}
	response.SetUser(user)
	return response, nil
}

func (client *testUserClient) GetUser(
	_ context.Context,
	request *userv1.GetUserRequest,
	_ ...grpc.CallOption,
) (*userv1.GetUserResponse, error) {
	client.getRequest = request
	user := &userv1.User{}
	user.SetId(request.GetId())
	response := &userv1.GetUserResponse{}
	response.SetUser(user)
	return response, nil
}

func TestUserRoutes(t *testing.T) {
	t.Parallel()
	client := &testUserClient{}
	mux := http.NewServeMux()
	user.RegisterHTTP(mux, client)

	create := httptest.NewRecorder()
	mux.ServeHTTP(
		create,
		httptest.NewRequest(
			http.MethodPost,
			"/v1/users",
			strings.NewReader(`{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`),
		),
	)
	require.Equal(t, http.StatusCreated, create.Code)
	require.Equal(t, "Ada", client.createRequest.GetFirstName())
	require.Equal(t, "Lovelace", client.createRequest.GetLastName())
	require.Equal(t, "ada@example.com", client.createRequest.GetEmail())
	require.JSONEq(
		t,
		`{"id":"user-id","first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`,
		create.Body.String(),
	)

	get := httptest.NewRecorder()
	mux.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/v1/users/user-id", nil))
	require.Equal(t, http.StatusOK, get.Code)
	require.Equal(t, "user-id", client.getRequest.GetId())

	invalid := httptest.NewRecorder()
	mux.ServeHTTP(
		invalid,
		httptest.NewRequest(
			http.MethodPost,
			"/v1/users",
			strings.NewReader(`{"first_name":"Ada","last_name":"Lovelace"}{}`),
		),
	)
	require.Equal(t, http.StatusBadRequest, invalid.Code)
}
