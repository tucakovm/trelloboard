package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"projects_module/domain"
	"projects_module/services"
)

type ProjectHandler struct {
	service services.ProjectService
}

func NewConnectionHandler(conns services.ProjectService) (ProjectHandler, error) {
	return ProjectHandler{
		service: conns,
	}, nil
}

func (c ProjectHandler) decodeBodyProject(r io.Reader) (*domain.Project, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt domain.Project
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (c *ProjectHandler) renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (h ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {

	project, err := h.decodeBodyProject(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.service.Create(*project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
