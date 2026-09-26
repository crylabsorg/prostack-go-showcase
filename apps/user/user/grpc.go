package user

import (
	"context"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"google.golang.org/grpc"
)

type server struct {
	userv1.UnimplementedUserServiceServer
	service *Service
}

func RegisterGRPC(grpcServer *grpc.Server, svc *Service) {
	userv1.RegisterUserServiceServer(grpcServer, &server{service: svc})
}

func (s *server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user, err := s.service.Create(ctx, req.GetName(), req.GetEmail())
	if err != nil {
		return nil, err
	}
	response := &userv1.CreateUserResponse{}
	response.SetUser(user.ToProto())
	return response, nil
}

func (s *server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	response := &userv1.GetUserResponse{}
	response.SetUser(user.ToProto())
	return response, nil
}
