package handlers

import (
	"net/http"
	"projects_module/services"
)

type ProjectHandler struct {
	conns services.ProjectService
}

func NewConnectionHandler(conns services.ProjectService) (ProjectHandler, error) {
	return ProjectHandler{
		conns: conns,
	}, nil
}

func (h ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {

}
