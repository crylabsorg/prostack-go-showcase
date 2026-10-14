package organization

import organizationv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/organization/v1"

type Organization struct {
	ID   string
	Name string
}

func (organization Organization) ToProto() *organizationv1.Organization {
	result := &organizationv1.Organization{}
	result.SetId(organization.ID)
	result.SetName(organization.Name)
	return result
}
