package models

type TaskStatusDuration struct {
	Status   string  `json:"status"`
	Duration float64 `json:"duration"`
}

type TaskDurations struct {
	TaskID          string               `json:"task_id"`
	StatusDurations []TaskStatusDuration `json:"status_durations"`
}

type TaskAssignments struct {
	Tasks []string `json:"tasks"`
}

type Analytic struct {
	ProjectID           string                     `json:"project_id"`
	TotalTasks          int32                      `json:"total_tasks"`
	StatusCounts        map[string]int32           `json:"status_counts"`
	TaskStatusDurations map[string]TaskDurations   `json:"task_status_durations"`
	MemberTasks         map[string]TaskAssignments `json:"member_tasks"`
	FinishedEarly       bool                       `json:"finished_early"`
}
