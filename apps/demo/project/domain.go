package project

import projectv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/project/v1"

type Project struct {
	ID             string
	OrganizationID string
	Name           string
}

func (project Project) ToProto() *projectv1.Project {
	result := &projectv1.Project{}
	result.SetId(project.ID)
	result.SetOrganizationId(project.OrganizationID)
	result.SetName(project.Name)
	return result
}
