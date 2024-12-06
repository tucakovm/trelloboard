package repository

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"workflow-service/models"
)

type WorkflowRepository struct {
	Driver neo4j.DriverWithContext
}

func NewWorkflowRepository(ctx context.Context) (*WorkflowRepository, error) {
	// Create a Neo4j driver
	driver, err := neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}
	// Return the repository with the driver
	return &WorkflowRepository{Driver: driver}, nil
}

func (r *WorkflowRepository) CreateWorkflow(ctx context.Context, workflow models.Workflow) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `CREATE (p:Project {id: $id, name: $name})`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":   workflow.ProjectID,
			"name": workflow.ProjectName,
		})
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) AddTask(ctx context.Context, projectID string, task models.TaskNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Add the task to the project
		taskQuery := `
			MATCH (p:Project {id: $projectID})
			CREATE (t:Task {id: $taskID, name: $taskName, blocked: false})
			CREATE (p)-[:HAS_TASK]->(t)`
		_, err := tx.Run(ctx, taskQuery, map[string]interface{}{
			"projectID": projectID,
			"taskID":    task.TaskID,
			"taskName":  task.TaskName,
		})
		if err != nil {
			return nil, err
		}

		// Create task dependencies
		for _, depID := range task.Dependencies {
			dependencyQuery := `
				MATCH (t:Task {id: $taskID}), (d:Task {id: $depID})
				CREATE (t)-[:DEPENDS_ON]->(d)`
			_, err = tx.Run(ctx, dependencyQuery, map[string]interface{}{
				"taskID": task.TaskID,
				"depID":  depID,
			})
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) GetTasks(ctx context.Context, projectID string) ([]models.TaskNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	tasks, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (p:Project {id: $projectID})-[:HAS_TASK]->(t:Task)
			OPTIONAL MATCH (t)-[:DEPENDS_ON]->(d:Task)
			RETURN t.id AS taskID, t.name AS taskName, COLLECT(d.id) AS dependencies`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"projectID": projectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		var tasks []models.TaskNode
		for result.Next(ctx) {
			record := result.Record()
			task := models.TaskNode{
				TaskID:       record.Values[0].(string),
				TaskName:     record.Values[1].(string),
				Dependencies: record.Values[2].([]string),
			}
			tasks = append(tasks, task)
		}
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("failed to read results: %w", err)
		}
		return tasks, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	return tasks.([]models.TaskNode), nil
}

func (r *WorkflowRepository) GetWorkflow(ctx context.Context, projectID string) (*models.Workflow, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	workflow, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (p:Project {id: $projectID}) RETURN p.id AS id, p.name AS name`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"projectID": projectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		if result.Next(ctx) {
			record := result.Record()
			return &models.Workflow{
				ProjectID:   record.Values[0].(string),
				ProjectName: record.Values[1].(string),
			}, nil
		}
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("failed to read results: %w", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	return workflow.(*models.Workflow), nil
}

func (r *WorkflowRepository) DeleteWorkflowByProjectID(ctx context.Context, projectID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Delete dependencies
		dependencyQuery := `
			MATCH (p:Project {id: $projectID})-[:HAS_TASK]->(t:Task)-[r:DEPENDS_ON]->(d:Task)
			DELETE r`
		_, err := tx.Run(ctx, dependencyQuery, map[string]interface{}{"projectID": projectID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete dependencies: %w", err)
		}

		// Delete tasks
		deleteTasksQuery := `
			MATCH (p:Project {id: $projectID})-[:HAS_TASK]->(t:Task)
			DETACH DELETE t`
		_, err = tx.Run(ctx, deleteTasksQuery, map[string]interface{}{"projectID": projectID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete tasks: %w", err)
		}

		// Delete the project node
		deleteProjectQuery := `
			MATCH (p:Project {id: $projectID})
			DETACH DELETE p`
		_, err = tx.Run(ctx, deleteProjectQuery, map[string]interface{}{"projectID": projectID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete project: %w", err)
		}

		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	return nil
}

func (r *WorkflowRepository) CheckTaskDependencies(ctx context.Context, projectID string, taskID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Execute the read transaction to check the dependencies
	_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Query to find all tasks that this task depends on
		query := `
			MATCH (t:Task {id: $taskID})-[:DEPENDS_ON]->(d:Task)
			RETURN d.id AS dependencyID, d.blocked AS blocked`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"taskID": taskID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		// Iterate over the dependencies and check if any are blocked
		for result.Next(ctx) {
			record := result.Record()
			blocked := record.Values[1].(bool)
			if blocked {
				// If any dependency is blocked, return false
				return false, nil
			}
		}
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("failed to read results: %w", err)
		}

		return true, nil
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
