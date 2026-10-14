package user_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crylabsorg/prostack-go/ps_error"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/crylabsorg/prostack-go-showcase/gateway/user"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

type testUserClient struct {
	userv1.UserServiceClient

	createRequest *userv1.CreateUserRequest
	getRequest    *userv1.GetUserRequest
	responseError error
}

func (client *testUserClient) CreateUser(
	_ context.Context,
	request *userv1.CreateUserRequest,
	_ ...grpc.CallOption,
) (*userv1.CreateUserResponse, error) {
	client.createRequest = request
	if client.responseError != nil {
		return nil, client.responseError
	}
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
	if client.responseError != nil {
		return nil, client.responseError
	}
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
	require.JSONEq(
		t,
		`{"id":"user-id","first_name":"","last_name":"","email":""}`,
		get.Body.String(),
	)
}

func TestCreateUserInvalidJSON(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"empty":          "",
		"malformed":      `{"first_name":`,
		"trailing value": `{"first_name":"Ada"}{}`,
		"unknown field":  `{"first_name":"Ada","unknown":true}`,
		"duplicate key":  `{"first_name":"Ada","first_name":"Grace"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &testUserClient{}
			mux := http.NewServeMux()
			user.RegisterHTTP(mux, client)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(body)))

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Equal(t, "text/plain; charset=utf-8", response.Header().Get("Content-Type"))
			require.Equal(t, "invalid JSON body\n", response.Body.String())
			require.Nil(t, client.createRequest)
		})
	}
}

func TestUserRoutesGRPCErrors(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name       string
		method     string
		path       string
		httpStatus int
		grpcStatus codes.Code
	}{
		{"invalid input", http.MethodPost, "/v1/users", http.StatusBadRequest, codes.InvalidArgument},
		{"email exists", http.MethodPost, "/v1/users", http.StatusConflict, codes.AlreadyExists},
		{"not found", http.MethodGet, "/v1/users/user-id", http.StatusNotFound, codes.NotFound},
		{"status from details", http.MethodGet, "/v1/users/user-id", http.StatusGone, codes.NotFound},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			client := &testUserClient{
				responseError: ps_error.PSError{
					ErrorCode: ps_error.ErrorCodeGeneral,
					Message:   "service error",
					Args:      []any{"email", 1},
					HStatus:   int32(testCase.httpStatus),
					GStatus:   testCase.grpcStatus,
				}.ToGRPCError(),
			}
			mux := http.NewServeMux()
			user.RegisterHTTP(mux, client)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(testCase.method, testCase.path, strings.NewReader(`{}`)))

			require.Equal(t, testCase.httpStatus, response.Code)
			require.Equal(t, "application/json", response.Header().Get("Content-Type"))
			require.JSONEq(t, `{"message":"service error","errorCode":0,"args":["email",1]}`, response.Body.String())
		})
	}
}

func TestUserRouteUnexpectedError(t *testing.T) {
	t.Parallel()
	client := &testUserClient{responseError: status.Error(codes.Internal, "unexpected error")}
	mux := http.NewServeMux()
	user.RegisterHTTP(mux, client)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/users/user-id", nil))

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), "unexpected error")
}
