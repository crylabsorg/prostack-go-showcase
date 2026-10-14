package organization

import (
	"encoding/json/v2"
	"net/http"

	"github.com/crylabsorg/prostack-go/controller"

	organizationv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/organization/v1"
)

func RegisterHTTP(mux *http.ServeMux, organizations organizationv1.OrganizationServiceClient) {
	handler := httpController{Controller: controller.NewController(), organizations: organizations}
	HandlerFromMux(handler, mux)
}

type httpController struct {
	controller.Controller

	organizations organizationv1.OrganizationServiceClient
}

func (c httpController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var input CreateOrganizationRequest
	if err := json.UnmarshalRead(r.Body, &input, json.RejectUnknownMembers(true)); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	result, err := c.organizations.CreateOrganization(r.Context(), input.ToProto())
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusCreated, newOrganizationResponse(result.GetOrganization()))
}

func (c httpController) GetOrganization(w http.ResponseWriter, r *http.Request, organizationID string) {
	request := &organizationv1.GetOrganizationRequest{}
	request.SetId(organizationID)
	result, err := c.organizations.GetOrganization(r.Context(), request)
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusOK, newOrganizationResponse(result.GetOrganization()))
}
