package organization

import (
	"context"

	"google.golang.org/grpc"

	organizationv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/organization/v1"
)

type server struct {
	organizationv1.UnimplementedOrganizationServiceServer

	service *Service
}

func RegisterGRPC(grpcServer *grpc.Server, service *Service) {
	organizationv1.RegisterOrganizationServiceServer(grpcServer, &server{service: service})
}

func (s *server) CreateOrganization(
	_ context.Context, request *organizationv1.CreateOrganizationRequest,
) (*organizationv1.CreateOrganizationResponse, error) {
	organization, err := s.service.Create(request.GetName())
	if err != nil {
		return nil, err
	}
	response := &organizationv1.CreateOrganizationResponse{}
	response.SetOrganization(organization.ToProto())
	return response, nil
}

func (s *server) GetOrganization(
	_ context.Context, request *organizationv1.GetOrganizationRequest,
) (*organizationv1.GetOrganizationResponse, error) {
	organization, err := s.service.Get(request.GetId())
	if err != nil {
		return nil, err
	}
	response := &organizationv1.GetOrganizationResponse{}
	response.SetOrganization(organization.ToProto())
	return response, nil
}
