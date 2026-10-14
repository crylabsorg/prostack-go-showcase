package project

import (
	"encoding/json/v2"
	"net/http"

	"github.com/crylabsorg/prostack-go/controller"

	projectv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/project/v1"
)

func RegisterHTTP(mux *http.ServeMux, projects projectv1.ProjectServiceClient) {
	handler := httpController{Controller: controller.NewController(), projects: projects}
	HandlerFromMux(handler, mux)
}

type httpController struct {
	controller.Controller

	projects projectv1.ProjectServiceClient
}

func (c httpController) CreateProject(w http.ResponseWriter, r *http.Request, organizationID string) {
	var input CreateProjectRequest
	if err := json.UnmarshalRead(r.Body, &input, json.RejectUnknownMembers(true)); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	result, err := c.projects.CreateProject(r.Context(), input.ToProto(organizationID))
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusCreated, newProjectResponse(result.GetProject()))
}

func (c httpController) GetProject(w http.ResponseWriter, r *http.Request, organizationID, projectID string) {
	request := &projectv1.GetProjectRequest{}
	request.SetOrganizationId(organizationID)
	request.SetId(projectID)
	result, err := c.projects.GetProject(r.Context(), request)
	if err != nil {
		c.WriteError(r.Context(), w, err)
		return
	}
	c.WriteJSON(r.Context(), w, http.StatusOK, newProjectResponse(result.GetProject()))
}
