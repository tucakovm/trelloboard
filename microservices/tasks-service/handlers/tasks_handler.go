package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"tasks-service/domain"
	proto "tasks-service/proto/task"
	"tasks-service/repository"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	//"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
)

type TaskHandler struct {
	repo *repository.TaskRepo // Use a pointer here
	proto.UnimplementedTaskServiceServer
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


func (h *TaskHandler) DeleteTask(ctx context.Context, req *proto.DeleteTaskReq) (*proto.EmptyResponse, error) {
    log.Printf("DeleteTask called with ID: %s", req.Id)

    id, err := uuid.Parse(req.Id)
    if err != nil {
        log.Printf("Invalid ID format: %s", req.Id)
        return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format")
    }

    if err := h.repo.Delete(id); err != nil {
        log.Printf("Failed to delete task with ID %s: %v", req.Id, err)
        return nil, status.Errorf(codes.Internal, "Failed to delete task")
    }

    log.Printf("Task with ID %s successfully deleted", req.Id)

    return &proto.EmptyResponse{}, nil
}



func (h *TaskHandler) CreateTask(ctx context.Context, req *proto.CreateTaskReq) (*proto.TaskResponse, error) {
    log.Println("CreateTask called with:", req)

    // Map the incoming gRPC request to your repository model
    newTask := domain.Task{
        Name:        req.Name,
        Description: req.Description,
       // DueDate:     req.DueDate.AsTime(), // Convert Timestamp to time.Time
        ProjectID:   req.ProjectId,
    }

    // Save the task using the repository
    createdTask, err := h.repo.Create(newTask)
    if err != nil {
        log.Printf("Failed to create task: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to create task")
    }

    return &proto.TaskResponse{
        Task: &proto.Task{
           // Id:          createdTask.ID,     
            Name:        createdTask.Name,
            Description: createdTask.Description,
           // DueDate:     timestamppb.New(createdTask),
            ProjectId:   createdTask.ProjectID,
        },
    }, nil
}

func (h *TaskHandler) GetAllByProjectID(ctx context.Context, req *proto.GetAllByProjectIDReq) (*proto.GetAllTasksRes, error) {
    log.Printf("GetAllByProjectID called with ProjectID: %s", req.ProjectId)

    tasks, err := h.repo.GetAllByProjectID(req.ProjectId)
    if err != nil {
        log.Printf("Failed to fetch tasks for ProjectID %s: %v", req.ProjectId, err)
        return nil, status.Errorf(codes.Internal, "Failed to fetch tasks")
    }

    log.Printf("Successfully fetched tasks for ProjectID %s", req.ProjectId)

    protoTasks := make([]*proto.Task, len(tasks))
    for i, task := range tasks {
        protoTasks[i] = &proto.Task{
            //Id:          task.ID.String(),
            Name:        task.Name,
            Description: task.Description,
           // DueDate:     timestamppb.New(task.DueDate),
            ProjectId:   task.ProjectID,
        }
    }

    return &proto.GetAllTasksRes{
        Tasks: protoTasks,
    }, nil
}


func (h *TaskHandler) GetAllTasks(ctx context.Context, req *proto.EmptyRequest) (*proto.GetAllTasksRes, error) {
    log.Println("GetAll called")

    // Call the repository to fetch all tasks
    tasks, err := h.repo.GetAll()
    if err != nil {
        log.Printf("Failed to fetch tasks: %v", err)
        return nil, status.Errorf(codes.Internal, "Failed to fetch tasks")
    }

    log.Println("Successfully fetched all tasks")

    // Convert repository tasks to proto tasks
    protoTasks := make([]*proto.Task, len(tasks))
    for i, task := range tasks {
        protoTasks[i] = &proto.Task{
            //Id:          task.ID.String(),
            Name:        task.Name,
            Description: task.Description,
           // DueDate:     timestamppb.New(task.DueDate),
            ProjectId:   task.ProjectID,
        }
    }

    // Return the response
    return &proto.GetAllTasksRes{
        Tasks: protoTasks,
    }, nil
}



//----------------------------------------------------
//ovo je bez grpc neka ga za sad za svaki slucaj 
//----------------------------------------------------




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

//func (h *TaskHandler) GetAllByProjectID(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	projectID := vars["project_id"]

	// Poziv repoza za dobijanje svih zadataka po projectID
//	tasks, err := h.repo.GetAllByProjectID(projectID)
//	if err != nil {
//		// Greška u pretrazi zadataka, vraća se status 500 i odgovarajuća poruka
//		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
//		return
//	}
//
//	log.Println("repo je prosao")
//	log.Println(tasks)
//	// Vraća uspešan odgovor sa statusom 200 i JSON podacima
//	h.renderJSON(w, tasks, http.StatusOK)
//}
