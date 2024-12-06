package models

type Workflow struct {
	ProjectID   string     `json:"project_id"`
	ProjectName string     `json:"project_name"`
	Tasks       []TaskNode `json:"tasks"`
}

type TaskNode struct {
	TaskID          string   `json:"task_id"`
	TaskName        string   `json:"task_name"`
	TaskDescription string   `json:"task_description"` // Task description
	Dependencies    []string `json:"dependencies"`     // IDs of dependent tasks
	Blocked         bool     `json:"blocked"`          // Whether the task is blocked
}
