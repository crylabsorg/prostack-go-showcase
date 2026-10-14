package project

import projectv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/project/v1"

func (request CreateProjectRequest) ToProto(organizationID string) *projectv1.CreateProjectRequest {
	result := &projectv1.CreateProjectRequest{}
	result.SetOrganizationId(organizationID)
	result.SetName(request.Name)
	return result
}

func newProjectResponse(project *projectv1.Project) ProjectResponse {
	return ProjectResponse{ID: project.GetId(), OrganizationID: project.GetOrganizationId(), Name: project.GetName()}
}
