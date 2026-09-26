package user

import (
	"context"
	"errors"

	userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, grpcError(err)
	}
	return &userv1.CreateUserResponse{User: toProto(user)}, nil
}

func (s *server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, grpcError(err)
	}
	return &userv1.GetUserResponse{User: toProto(user)}, nil
}

func toProto(user User) *userv1.User {
	return &userv1.User{Id: user.ID, Name: user.Name, Email: user.Email}
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrEmailExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
