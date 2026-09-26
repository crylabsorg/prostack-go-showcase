package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type testUserClient struct {
	userv1.UserServiceClient
	createRequest *userv1.CreateUserRequest
	getRequest    *userv1.GetUserRequest
}

func (client *testUserClient) CreateUser(_ context.Context, request *userv1.CreateUserRequest, _ ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	client.createRequest = request
	user := &userv1.User{}
	user.SetId("user-id")
	user.SetName(request.GetName())
	user.SetEmail(request.GetEmail())
	response := &userv1.CreateUserResponse{}
	response.SetUser(user)
	return response, nil
}

func (client *testUserClient) GetUser(_ context.Context, request *userv1.GetUserRequest, _ ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	client.getRequest = request
	user := &userv1.User{}
	user.SetId(request.GetId())
	response := &userv1.GetUserResponse{}
	response.SetUser(user)
	return response, nil
}

func TestUserRoutes(t *testing.T) {
	client := &testUserClient{}
	mux := http.NewServeMux()
	RegisterHTTP(mux, client)

	create := httptest.NewRecorder()
	mux.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(`{"name":"Ada","email":"ada@example.com"}`)))
	require.Equal(t, http.StatusCreated, create.Code)
	require.Equal(t, "Ada", client.createRequest.GetName())
	require.Equal(t, "ada@example.com", client.createRequest.GetEmail())
	require.JSONEq(t, `{"id":"user-id","name":"Ada","email":"ada@example.com"}`, create.Body.String())

	get := httptest.NewRecorder()
	mux.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/v1/users/user-id", nil))
	require.Equal(t, http.StatusOK, get.Code)
	require.Equal(t, "user-id", client.getRequest.GetId())

	invalid := httptest.NewRecorder()
	mux.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(`{"name":"Ada"}{}`)))
	require.Equal(t, http.StatusBadRequest, invalid.Code)
}
