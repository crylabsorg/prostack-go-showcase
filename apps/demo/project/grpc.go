package project

import (
	"context"

	"google.golang.org/grpc"

	projectv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/project/v1"
)

type server struct {
	projectv1.UnimplementedProjectServiceServer

	service *Service
}

func RegisterGRPC(grpcServer *grpc.Server, service *Service) {
	projectv1.RegisterProjectServiceServer(grpcServer, &server{service: service})
}

func (s *server) CreateProject(
	_ context.Context, request *projectv1.CreateProjectRequest,
) (*projectv1.CreateProjectResponse, error) {
	project, err := s.service.Create(request.GetOrganizationId(), request.GetName())
	if err != nil {
		return nil, err
	}
	response := &projectv1.CreateProjectResponse{}
	response.SetProject(project.ToProto())
	return response, nil
}

func (s *server) GetProject(
	_ context.Context, request *projectv1.GetProjectRequest,
) (*projectv1.GetProjectResponse, error) {
	project, err := s.service.Get(request.GetOrganizationId(), request.GetId())
	if err != nil {
		return nil, err
	}
	response := &projectv1.GetProjectResponse{}
	response.SetProject(project.ToProto())
	return response, nil
}
