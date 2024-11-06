package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"tasks-service/domain"
	"tasks-service/service"
)

type TaskHandler struct {
	service service.TaskService
}

func NewConnectionHandler(conn service.TaskService) (TaskHandler, error) {
	return TaskHandler{service: conn}, nil
}

func (c TaskHandler) decodeBodyProject(r io.Reader) (*domain.Task, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt domain.Task
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (c *TaskHandler) renderJSON(w http.ResponseWriter, v interface{}, code int) {
	js, err := json.Marshal(v)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

func (h TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	task, err := h.decodeBodyProject(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.service.Create(*task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.renderJSON(w, t, http.StatusCreated)

}
