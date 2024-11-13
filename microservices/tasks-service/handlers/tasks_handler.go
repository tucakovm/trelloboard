package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"tasks-service/domain"
	"tasks-service/repository"

	"github.com/gorilla/mux"

	"github.com/google/uuid"
)

type TaskHandler struct {
	repo *repository.TaskRepo // Use a pointer here
}

func NewTaskHandler(repo *repository.TaskRepo) *TaskHandler { // Accept pointer in constructor
	return &TaskHandler{repo: repo}
}

func (h *TaskHandler) decodeBodyTask(r io.Reader) (*domain.Task, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var task domain.Task
	if err := dec.Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (h *TaskHandler) renderJSON(w http.ResponseWriter, v interface{}, code int) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.Println("aaaaaaaa")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		log.Println("options proso")
		return
	}
	log.Println("proso options")

	task, err := h.decodeBodyTask(r.Body)
	if err != nil {
		log.Println("greska u dekodiranju bodya")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("dekodiran body")

	//task.Id = uuid.New()

	createdTask, err := h.repo.Create(*task)
	log.Println("pokrenut create task u handleru")
	if err != nil {
		log.Println("greska u kreairanju taska u repo")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("ide gas")

	h.renderJSON(w, createdTask, http.StatusCreated)
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, tasks, http.StatusOK)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) DeleteAllByProjectID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	if err := h.repo.DeleteAllByProjectID(projectID); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
