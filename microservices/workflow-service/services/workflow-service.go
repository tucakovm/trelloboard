package services

import (
	"context"
	"fmt"
	"workflow-service/models"
	"workflow-service/repository"
)

type WorkflowService interface {
	CreateWorkflow(workflow models.Workflow) error
	AddTask(projectID string, task models.TaskNode) error
	GetTasks(projectID string) ([]models.TaskNode, error)
	GetWorkflow(projectID string) (*models.Workflow, error)
	DeleteWorkflowByProjectID(projectID string) error
	CheckTaskDependencies(projectID string, taskID string) (bool, error)
}

type workflowService struct {
	repo repository.WorkflowRepository
}

// NewWorkflowService creates a new instance of WorkflowService
func NewWorkflowService(repo repository.WorkflowRepository) WorkflowService {
	return &workflowService{repo: repo}
}

// CreateWorkflow creates a new workflow
func (s *workflowService) CreateWorkflow(workflow models.Workflow) error {
	ctx := context.Background()
	// Ensure proper error handling when calling the repository method
	if err := s.repo.CreateWorkflow(ctx, workflow); err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}
	return nil
}

// AddTask adds a new task to an existing workflow
func (s *workflowService) AddTask(projectID string, task models.TaskNode) error {
	ctx := context.Background()
	// Ensure proper error handling when calling the repository method
	if err := s.repo.AddTask(ctx, projectID, task); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}

// GetTasks retrieves all tasks for a given project ID
func (s *workflowService) GetTasks(projectID string) ([]models.TaskNode, error) {
	ctx := context.Background()
	tasks, err := s.repo.GetTasks(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for project %s: %w", projectID, err)
	}
	return tasks, nil
}

// GetWorkflow retrieves the workflow for a given project ID
func (s *workflowService) GetWorkflow(projectID string) (*models.Workflow, error) {
	ctx := context.Background()
	workflow, err := s.repo.GetWorkflow(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow for project %s: %w", projectID, err)
	}
	return workflow, nil
}

// DeleteWorkflowByProjectID deletes a workflow for a given project ID
func (s *workflowService) DeleteWorkflowByProjectID(projectID string) error {
	ctx := context.Background()
	if err := s.repo.DeleteWorkflowByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete workflow for project %s: %w", projectID, err)
	}
	return nil
}

// CheckTaskDependencies checks if all dependencies for a task are met
func (s *workflowService) CheckTaskDependencies(projectID string, taskID string) (bool, error) {
	// Fetch the workflow for the project
	workflow, err := s.repo.GetWorkflow(context.Background(), projectID)
	if err != nil {
		return false, fmt.Errorf("failed to get workflow for project %s: %w", projectID, err)
	}

	// Find the task by ID
	var task *models.TaskNode
	for _, t := range workflow.Tasks {
		if t.TaskID == taskID {
			task = &t
			break
		}
	}
	if task == nil {
		return false, fmt.Errorf("task %s not found in project %s", taskID, projectID)
	}

	// Check if all dependencies are met
	for _, depID := range task.Dependencies {
		// Check if the dependency exists in the workflow
		dependencyMet := false
		for _, t := range workflow.Tasks {
			if t.TaskID == depID {
				dependencyMet = true
				break
			}
		}
		if !dependencyMet {
			return false, fmt.Errorf("dependency %s not found for task %s", depID, taskID)
		}
	}

	// All dependencies are met
	return true, nil
}