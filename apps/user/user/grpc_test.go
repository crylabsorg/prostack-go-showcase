package user_test

import (
	"context"
	"encoding/json/v2"
	"net"
	"path/filepath"
	"testing"

	"github.com/crylabsorg/prostack-go/ps_test"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	usersvc "github.com/crylabsorg/prostack-go-showcase/user/user"
)

func TestGRPCOptionalNames(t *testing.T) {
	t.Parallel()
	pool := ps_test.NewPostgresContainer(t, filepath.Join("..", "schema.sql"))
	repository := usersvc.NewRepository(pool)
	const bufferSize = 1024 * 1024
	listener := bufconn.Listen(bufferSize)
	grpcServer := grpc.NewServer()
	usersvc.RegisterGRPC(grpcServer, usersvc.NewService(repository, usersvc.NewMetrics(prometheus.NewRegistry())))
	serveResult := make(chan error, 1)
	go func() { serveResult <- grpcServer.Serve(listener) }()
	t.Cleanup(func() { grpcServer.Stop(); require.NoError(t, <-serveResult) })
	connection, connectionErr := grpc.NewClient(
		"passthrough:///user",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
	)
	require.NoError(t, connectionErr)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	handler := userv1.NewUserServiceClient(connection)
	for _, testCase := range []struct {
		name      string
		firstName *string
		lastName  *string
	}{
		{name: "omitted"},
		{name: "empty", firstName: new(""), lastName: new("")},
		{name: "first", firstName: new("Ada")},
		{name: "last", lastName: new("Lovelace")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			request := userv1.CreateUserRequest_builder{
				Email:     new(testCase.name + "@example.com"),
				FirstName: testCase.firstName,
				LastName:  testCase.lastName,
			}.Build()
			created, err := handler.CreateUser(t.Context(), request)
			require.NoError(t, err)
			result, err := handler.GetUser(
				t.Context(),
				userv1.GetUserRequest_builder{Id: new(created.GetUser().GetId())}.Build(),
			)
			require.NoError(t, err)
			require.True(t, proto.Equal(created.GetUser(), result.GetUser()))
			require.Equal(t, testCase.firstName != nil, result.GetUser().HasFirstName())
			require.Equal(t, testCase.lastName != nil, result.GetUser().HasLastName())
			var storedFirstName, storedLastName *string
			require.NoError(
				t,
				pool.QueryRow(t.Context(), "SELECT first_name,last_name FROM users WHERE id=$1", result.GetUser().GetId()).
					Scan(&storedFirstName, &storedLastName),
			)
			require.Equal(t, testCase.firstName, storedFirstName)
			require.Equal(t, testCase.lastName, storedLastName)
			var payload []byte
			require.NoError(
				t,
				pool.QueryRow(t.Context(), "SELECT payload FROM user_outbox WHERE user_id=$1", result.GetUser().GetId()).
					Scan(&payload),
			)
			var event events.UserCreatedEvent
			require.NoError(t, json.Unmarshal(payload, &event))
			require.Equal(t, testCase.firstName, event.FirstName)
			require.Equal(t, testCase.lastName, event.LastName)
			var members map[string]any
			require.NoError(t, json.Unmarshal(payload, &members))
			_, hasFirstName := members["first_name"]
			_, hasLastName := members["last_name"]
			require.Equal(t, testCase.firstName != nil, hasFirstName)
			require.Equal(t, testCase.lastName != nil, hasLastName)
		})
	}
}
