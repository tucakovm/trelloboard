package repository

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"log"
	"workflow-service/models"
)

type WorkflowRepository struct {
	Driver neo4j.DriverWithContext
}

func NewWorkflowRepository(ctx context.Context) (*WorkflowRepository, error) {
	// Create a Neo4j driver once
	driver, err := neo4j.NewDriverWithContext("bolt://neo4j:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}
	// Return the repository with the driver
	return &WorkflowRepository{Driver: driver}, nil
}

func (r *WorkflowRepository) CreateWorkflow(ctx context.Context, workflow models.Workflow) error {
	log.Printf("Creating workflow in database: project_id=%s, project_name=%s", workflow.ProjectID, workflow.ProjectName)

	// Create session for this transaction
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Provera da li postoji projekat sa istim imenom
		checkQuery := `MATCH (p:Project {name: $name}) RETURN p`
		result, err := tx.Run(ctx, checkQuery, map[string]interface{}{
			"name": workflow.ProjectName,
		})
		if err != nil {
			return nil, err
		}

		// Ako postoji projekat sa istim imenom, vratimo grešku
		if result.Next(ctx) {
			return nil, fmt.Errorf("workflow with name '%s' already exists", workflow.ProjectName)
		}

		// Ako ne postoji, kreiramo novi projekat
		createQuery := `CREATE (p:Project {id: $id, name: $name})`
		_, err = tx.Run(ctx, createQuery, map[string]interface{}{
			"id":   workflow.ProjectID,
			"name": workflow.ProjectName,
		})
		return nil, err
	})
	if err != nil {
		log.Printf("Error creating workflow in Neo4j: %v", err)
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	log.Printf("Workflow created successfully in Neo4j: project_id=%s", workflow.ProjectID)

	return nil
}

func (r *WorkflowRepository) AddTask(ctx context.Context, projectID string, task models.TaskNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("Checking if task exists for projectID: %s, taskID: %s", projectID, task.TaskID)

		checkTaskQuery := `MATCH (t:Task {id: $taskID}) RETURN t`
		result, err := tx.Run(ctx, checkTaskQuery, map[string]interface{}{
			"taskID": task.TaskID,
		})
		if err != nil {
			return nil, err
		}

		taskExists := result.Next(ctx)

		if !taskExists {
			log.Printf("Task %s does not exist, creating new task", task.TaskID)
			taskQuery := `
                MATCH (p:Project {id: $projectID}) 
                CREATE (t:Task {id: $taskID, name: $taskName, description: $taskDescription, blocked: $taskBlocked}) 
                CREATE (p)-[:HAS_TASK]->(t)`
			_, err := tx.Run(ctx, taskQuery, map[string]interface{}{
				"projectID":       projectID,
				"taskID":          task.TaskID,
				"taskName":        task.TaskName,
				"taskDescription": task.TaskDescription,
				"taskBlocked":     task.Blocked,
			})
			if err != nil {
				return nil, err
			}
		} else {
			log.Printf("Task %s already exists, skipping creation", task.TaskID)
		}

		log.Printf("Dependencies for task %s: %v", task.TaskID, task.Dependencies)

		// Dodavanje zavisnosti
		for _, depID := range task.Dependencies {
			checkDependencyQuery := `MATCH (d:Task {id: $depID}) RETURN d`
			result, err := tx.Run(ctx, checkDependencyQuery, map[string]interface{}{
				"depID": depID,
			})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				checkCycleQuery := `
                    MATCH (t:Task {id: $taskID}), (d:Task {id: $depID})
                    WITH t, d
                    MATCH p = shortestPath((d)-[:DEPENDS_ON*]->(t))
                    RETURN p`
				cycleResult, err := tx.Run(ctx, checkCycleQuery, map[string]interface{}{
					"taskID": task.TaskID,
					"depID":  depID,
				})
				if err != nil {
					return nil, err
				}
				if cycleResult.Next(ctx) {
					log.Printf("Cycle detected: Task %s depends on Task %s, which would create a cycle", task.TaskID, depID)
					return nil, fmt.Errorf("adding dependency would create a cycle: task %s depends on task %s", task.TaskID, depID)
				}

				dependencyQuery := `
                    MATCH (t:Task {id: $taskID}), (d:Task {id: $depID})
                    MERGE (t)-[:DEPENDS_ON]->(d)`
				_, err = tx.Run(ctx, dependencyQuery, map[string]interface{}{
					"taskID": task.TaskID,
					"depID":  depID,
				})
				if err != nil {
					return nil, err
				}
			} else {
				log.Printf("Dependency task with ID %s does not exist", depID)
				return nil, fmt.Errorf("dependency task with ID %s does not exist", depID)
			}
		}
		return nil, nil
	})
	if err != nil {
		log.Printf("Error while adding task: %v", err)
		return fmt.Errorf("failed to add task: %w", err)
	}

	log.Printf("Task processing completed for projectID: %s, taskID: %s", projectID, task.TaskID)
	return nil
}

func (r *WorkflowRepository) GetTasks(ctx context.Context, projectID string) ([]models.TaskNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	tasks, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (p:Project {id: $projectID})-[:HAS_TASK]->(t:Task)
			OPTIONAL MATCH (t)-[:DEPENDS_ON]->(d:Task)
			RETURN t.id AS taskID, 
			       t.name AS taskName, 
			       t.description AS taskDescription, 
			       t.blocked AS taskBlocked, 
			       COLLECT(d.id) AS dependencies`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"projectID": projectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		var tasks []models.TaskNode
		for result.Next(ctx) {
			record := result.Record()

			dependenciesRaw := record.Values[4].([]interface{})
			dependencies := make([]string, len(dependenciesRaw))
			for i, dep := range dependenciesRaw {
				dependencies[i] = dep.(string)
			}

			task := models.TaskNode{
				TaskID:          record.Values[0].(string),
				TaskName:        record.Values[1].(string),
				TaskDescription: record.Values[2].(string),
				Blocked:         record.Values[3].(bool),
				Dependencies:    dependencies,
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

	log.Printf("usao u repo neo4j")
	workflowAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
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
				Tasks:       nil,
			}, nil
		}

		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("failed to read results: %w", err)
		}

		return nil, nil
	})
	log.Printf("usao2 u repo neo4j %s", workflowAny)
	if workflowAny == nil {
		return &models.Workflow{
			ProjectID:   projectID,
			ProjectName: "",
			Tasks:       []models.TaskNode{},
		}, nil
	}

	workflow, ok := workflowAny.(*models.Workflow)
	if !ok {
		return nil, fmt.Errorf("unexpected type assertion failure: got %T", workflowAny)
	}

	tasks, err := r.GetTasks(ctx, projectID)
	if err != nil {
		log.Printf("Workflow repo task getTasks error: %s", err)
		return nil, fmt.Errorf("failed to get tasks for workflow: %w", err)
	}

	workflow.Tasks = tasks
	log.Printf("Workflow retrieved for projectID %s with %d tasks", projectID, len(tasks))

	return workflow, nil
}

func (r *WorkflowRepository) DeleteWorkflowByProjectID(ctx context.Context, projectID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	log.Printf("dosao do delete workflow: %s\n", projectID)
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

func (r *WorkflowRepository) CheckProjectsExist(ctx context.Context) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	exists, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := "MATCH (p:Project) RETURN COUNT(p) > 0 AS exists"
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		if result.Next(ctx) {
			record := result.Record()
			return record.Values[0].(bool), nil
		}
		return false, result.Err()
	})

	if err != nil {
		return false, err
	}

	return exists.(bool), nil
}

func checkWorkflowsExist(ctx context.Context, session neo4j.Session) (bool, error) {
	query := "MATCH (w:Workflow) RETURN COUNT(w) > 0 AS exists"
	result, err := session.Run(query, nil)
	if err != nil {
		return false, err
	}

	if result.Next() {
		exists, _ := result.Record().Get("exists")
		return exists.(bool), nil
	}

	return false, result.Err()
}

// Funkcija za dobavljanje svih taskova iz svih workflow-ova
func (r *WorkflowRepository) GetAllTasksFromAllWorkflows(ctx context.Context) ([]models.TaskNode, error) {
	// Otvaranje sesije
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Cypher upit za dobijanje svih taskova
	query := `
		MATCH (:Workflow)-[:HAS_TASK]->(t:Task)
		RETURN t.id AS id, t.name AS name, t.description AS description, 
		t.dependencies AS dependencies, t.blocked AS blocked
	`

	tasks := []models.TaskNode{}

	// Izvršavanje upita
	_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		for records.Next(ctx) {
			record := records.Record()

			dependenciesRaw := record.Values[3].([]interface{})
			dependencies := make([]string, len(dependenciesRaw))
			for i, dep := range dependenciesRaw {
				dependencies[i] = dep.(string)
			}

			task := models.TaskNode{
				TaskID:          record.Values[0].(string),
				TaskName:        record.Values[1].(string),
				TaskDescription: record.Values[2].(string),
				Dependencies:    dependencies,
				Blocked:         record.Values[4].(bool),
			}
			tasks = append(tasks, task)
		}

		return nil, records.Err()
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// Provera da li zadati ID postoji među svim taskovima
func (r *WorkflowRepository) TaskExistsInAllWorkflows(ctx context.Context, taskID string) (bool, error) {
	tasks, err := r.GetAllTasksFromAllWorkflows(ctx)
	log.Printf("tasks: %w", tasks)

	if err != nil {
		log.Printf("failed to get tasks: %w", err)

		return false, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Iteracija kroz taskove radi provere ID-a
	for _, task := range tasks {
		log.Printf("task: %w", task.TaskID)

		if task.TaskID == taskID {
			return true, nil
		}
	}
	return false, nil
}

func (r *WorkflowRepository) IsTaskBlocked(ctx context.Context, taskID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	blocked, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (t:Task {id: $taskID}) RETURN t.blocked AS blocked`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"taskID": taskID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		if result.Next(ctx) {
			record := result.Record()
			blockedValue, ok := record.Values[0].(bool)
			if !ok {
				return nil, fmt.Errorf("unexpected type for blocked field")
			}
			return blockedValue, nil
		}

		return false, nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if task is blocked: %w", err)
	}

	boolValue, ok := blocked.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for blocked field")
	}

	log.Printf("blocked, %t: ", boolValue)
	return boolValue, nil
}

func (r *WorkflowRepository) UpdateTaskStatus(ctx context.Context, taskID string, blocked bool) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("Updating status for task %s to blocked=%v", taskID, blocked)

		updateQuery := `
			MATCH (t:Task {id: $taskID})
			SET t.blocked = $blocked
			RETURN t`
		result, err := tx.Run(ctx, updateQuery, map[string]interface{}{
			"taskID":  taskID,
			"blocked": blocked,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run update query: %w", err)
		}

		if !result.Next(ctx) {
			return nil, fmt.Errorf("task with ID %s not found", taskID)
		}

		err = r.UpdateBlockedStatusForDependents(ctx, taskID)
		if err != nil {
			return nil, err
		}

		log.Printf("Task %s status successfully updated", taskID)
		return nil, nil
	})

	if err != nil {
		log.Printf("Error updating task status: %v", err)
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) IsTaskBlockedByDependency(ctx context.Context, taskID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (t:Task {id: $taskID})-[:DEPENDS_ON]->(d:Task)
			WHERE d.blocked = true
			RETURN COUNT(d) > 0 AS isBlockedByDependency`

		res, err := tx.Run(ctx, query, map[string]interface{}{
			"taskID": taskID,
		})
		if err != nil {
			return false, fmt.Errorf("query execution failed: %w", err)
		}

		if res.Next(ctx) {
			return res.Record().Values[0].(bool), nil
		}

		if err := res.Err(); err != nil {
			return false, fmt.Errorf("result reading failed: %w", err)
		}

		return false, nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to check blocked dependency: %w", err)
	}
	return result.(bool), nil
}

func (r *WorkflowRepository) UpdateBlockedStatusForDependents(ctx context.Context, taskID string) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return r.updateDependentsRecursive(ctx, tx, taskID)
	})

	if err != nil {
		log.Printf("Error while updating blocked statuses: %v", err)
		return fmt.Errorf("failed to update blocked statuses: %w", err)
	}

	return nil
}

func (r *WorkflowRepository) updateDependentsRecursive(ctx context.Context, tx neo4j.ManagedTransaction, taskID string) (interface{}, error) {
	query := `
		MATCH (t:Task {id: $taskID})<-[:DEPENDS_ON]-(dependent:Task)
		RETURN dependent.id AS dependentID, dependent.blocked AS currentBlocked`

	result, err := tx.Run(ctx, query, map[string]interface{}{
		"taskID": taskID,
	})
	if err != nil {
		return nil, err
	}

	for result.Next(ctx) {
		record := result.Record()
		dependentID, _ := record.Get("dependentID")
		currentBlocked, _ := record.Get("currentBlocked")

		checkDepsQuery := `
			MATCH (dependent:Task {id: $dependentID})-[:DEPENDS_ON]->(dep:Task)
			RETURN dep.blocked AS blocked`

		depResult, err := tx.Run(ctx, checkDepsQuery, map[string]interface{}{
			"dependentID": dependentID,
		})
		if err != nil {
			return nil, err
		}

		newStatus := true

		for depResult.Next(ctx) {
			blockedVal, _ := depResult.Record().Get("blocked")
			if val, ok := blockedVal.(bool); ok && val {
				continue
			}
			newStatus = false
			break
		}

		if currentBlockedBool, ok := currentBlocked.(bool); ok && currentBlockedBool != newStatus {
			updateQuery := `
				MATCH (t:Task {id: $dependentID})
				SET t.blocked = $newStatus`

			_, err = tx.Run(ctx, updateQuery, map[string]interface{}{
				"dependentID": dependentID,
				"newStatus":   newStatus,
			})
			if err != nil {
				return nil, err
			}

			log.Printf("Updated blocked status for task %s to %v", dependentID, newStatus)

			_, err = r.updateDependentsRecursive(ctx, tx, dependentID.(string))
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
