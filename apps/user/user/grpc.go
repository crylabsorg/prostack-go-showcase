package user

import (
	"context"

	"google.golang.org/grpc"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
)

type server struct {
	userv1.UnimplementedUserServiceServer

	service *Service
}

func RegisterGRPC(grpcServer *grpc.Server, svc *Service) {
	userv1.RegisterUserServiceServer(grpcServer, &server{service: svc})
}

func (s *server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	var firstName, lastName *string
	if req.HasFirstName() {
		firstName = new(req.GetFirstName())
	}
	if req.HasLastName() {
		lastName = new(req.GetLastName())
	}
	user, err := s.service.Create(ctx, firstName, lastName, req.GetEmail())
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
