package handlers

import (
	"context"
	"log"
	"workflow-service/models"
	proto "workflow-service/proto/workflows" // Import your generated proto package
	"workflow-service/services"
)

type WorkflowHandler struct {
	WorkflowService services.WorkflowService
	proto.UnimplementedWorkflowServiceServer
}

func NewWorkflowHandler(service services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{WorkflowService: service}
}

// Implement the gRPC method for CreateWorkflow
func (h *WorkflowHandler) CreateWorkflow(ctx context.Context, req *proto.CreateWorkflowReq) (*proto.VoidResponse, error) {
	log.Printf("Received CreateWorkflow request: project_id=%s, project_name=%s", req.ProjectId, req.ProjectName)

	// Convert the proto request to your models
	workflow := models.Workflow{
		ProjectID:   req.ProjectId,
		ProjectName: req.ProjectName,
		Tasks:       []models.TaskNode{},
	}

	// Call the service to create the workflow
	if err := h.WorkflowService.CreateWorkflow(workflow); err != nil {
		return nil, err
	}

	log.Printf("Workflow created successfully for project_id=%s", req.ProjectId)
	return &proto.VoidResponse{}, nil
}

// Implement the gRPC method for AddTask
func (h *WorkflowHandler) AddTask(ctx context.Context, req *proto.AddTaskReq) (*proto.VoidResponse, error) {

	projectID := req.ProjectId
	task := models.TaskNode{
		TaskID:   req.Task.Id,
		TaskName: req.Task.Name,

		Dependencies: req.Task.Dependencies,
		Blocked:      req.Task.Blocked, // Include Blocked field
	}

	log.Printf("Handler for task %s with projectId: %v", task.TaskID, projectID)

	// Call the service to add the task
	if err := h.WorkflowService.AddTask(projectID, task); err != nil {
		return nil, err
	}

	return &proto.VoidResponse{}, nil
}

// Implement the gRPC method for GetWorkflowByProjectID
func (h *WorkflowHandler) GetWorkflowByProjectID(ctx context.Context, req *proto.GetWorkflowReq) (*proto.GetWorkflowRes, error) {
	projectID := req.ProjectId

	// Get the workflow from the service
	workflow, err := h.WorkflowService.GetWorkflow(projectID)
	if err != nil {
		return nil, err
	}

	// Convert the workflow to the proto response format
	taskResponses := make([]*proto.Task, len(workflow.Tasks))
	for i, task := range workflow.Tasks {
		taskResponses[i] = &proto.Task{
			Id:           task.TaskID,
			Name:         task.TaskName,
			Dependencies: task.Dependencies,
			Blocked:      task.Blocked, // Include Blocked field
		}
	}

	return &proto.GetWorkflowRes{
		Workflow: &proto.Workflow{
			ProjectId:   workflow.ProjectID,
			ProjectName: workflow.ProjectName,
			Tasks:       taskResponses,
		},
	}, nil
}

// Implement the gRPC method for DeleteWorkflowByProjectID
func (h *WorkflowHandler) DeleteWorkflowByProjectID(ctx context.Context, req *proto.GetWorkflowReq) (*proto.VoidResponse, error) {
	projectID := req.ProjectId

	// Call the service to delete the workflow
	if err := h.WorkflowService.DeleteWorkflowByProjectID(projectID); err != nil {
		return nil, err
	}

	return &proto.VoidResponse{}, nil
}

// Implement the gRPC method for CheckTaskDependencies
func (h *WorkflowHandler) CheckTaskDependencies(ctx context.Context, req *proto.CheckTaskDependenciesReq) (*proto.TaskDependenciesStatus, error) {
	projectID := req.ProjectId
	taskID := req.TaskId

	// Check task dependencies in the service
	allDependenciesMet, err := h.WorkflowService.CheckTaskDependencies(projectID, taskID)
	if err != nil {
		return nil, err
	}

	return &proto.TaskDependenciesStatus{AllDependenciesMet: allDependenciesMet}, nil
}
