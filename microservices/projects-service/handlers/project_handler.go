package handlers

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"projects_module/domain"
	"projects_module/services"
)

type KeyProduct struct{}

type ProjectHandler struct {
	service services.ProjectService
}

func NewConnectionHandler(service services.ProjectService) (ProjectHandler, error) {
	return ProjectHandler{
		service: service,
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

func (c ProjectHandler) renderJSON(w http.ResponseWriter, v interface{}, code int) {
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
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	//project, err := h.decodeBodyProject(r.Body)
	project := r.Context().Value(KeyProduct{}).(*domain.Project)
	log.Println(project)

	if project.Manager.Role != "Manager" {
		http.Error(w, "", http.StatusUnauthorized)
	}

	h.service.Create(project)

	h.renderJSON(w, domain.Projects{}, http.StatusCreated)
}

func (h ProjectHandler) GetAll(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	// Call the service to get all projects
	allProducts, err := h.service.GetAll()
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		return
	}

	// Marshal and write the response
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(allProducts)
	if err != nil {
		http.Error(rw, "Error marshalling data", http.StatusInternalServerError)
		return
	}
	rw.Write(jsonData)
}

func (h ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.Delete(id)
	if err != nil {
		http.Error(w, "Failed to delete config", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, "Project deleted", http.StatusOK)
}

func (h ProjectHandler) MiddlewareProjectDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		project := &domain.Project{}
		err := project.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			log.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, project)
		h = h.WithContext(ctx)

		next.ServeHTTP(rw, h)
	})
}
