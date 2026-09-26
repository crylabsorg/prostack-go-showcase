package user

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crylabsorg/prostack-go/controller"
	"github.com/crylabsorg/prostack-go/ps_error"
	"github.com/crylabsorg/prostack-go/ps_grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceRejectsInvalidInput(t *testing.T) {
	svc := &Service{}
	for _, call := range []func() error{
		func() error { _, err := svc.Create(context.Background(), "Ada", "not-an-email"); return err },
		func() error { _, err := svc.Get(context.Background(), "not-a-uuid"); return err },
	} {
		var got ps_error.PSError
		require.ErrorAs(t, call(), &got)
		require.Equal(t, ErrInvalidInput, got)
	}
}

func TestUserErrorHTTPStatus(t *testing.T) {
	for _, tc := range []struct {
		err        ps_error.PSError
		httpStatus int
		grpcStatus codes.Code
	}{
		{ErrInvalidInput, http.StatusBadRequest, codes.InvalidArgument},
		{ErrNotFound, http.StatusNotFound, codes.NotFound},
		{ErrEmailExists, http.StatusConflict, codes.AlreadyExists},
	} {
		t.Run(tc.err.Message, func(t *testing.T) {
			_, rpcErr := ps_grpc.ErrorInterceptor(context.Background(), nil, nil, func(context.Context, any) (any, error) {
				return nil, fmt.Errorf("wrapped: %w", tc.err)
			})
			require.Equal(t, tc.grpcStatus, status.Code(rpcErr))
			response := httptest.NewRecorder()
			controller.NewController().WriteError(context.Background(), response, rpcErr)
			require.Equal(t, tc.httpStatus, response.Code)
		})
	}
}
