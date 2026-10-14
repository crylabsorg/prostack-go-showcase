package organization

import organizationv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/organization/v1"

func (request CreateOrganizationRequest) ToProto() *organizationv1.CreateOrganizationRequest {
	result := &organizationv1.CreateOrganizationRequest{}
	result.SetName(request.Name)
	return result
}

func newOrganizationResponse(organization *organizationv1.Organization) OrganizationResponse {
	return OrganizationResponse{ID: organization.GetId(), Name: organization.GetName()}
}
