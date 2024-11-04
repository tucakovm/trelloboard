package handlers

import (
	"encoding/json"
	"io"
	"log"
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

func (c *ProjectHandler) renderJSON(w http.ResponseWriter, v interface{}, code int) {
	js, err := json.Marshal(v)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

func (h ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {

	project, err := h.decodeBodyProject(r.Body)
	log.Println(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prj, err := h.service.Create(*project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.renderJSON(w, prj, http.StatusCreated)
	log.Println(prj)
}
