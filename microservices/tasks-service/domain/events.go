package domain

import "time"

type ProjectEvent struct {
	ProjectID       string    `json:"project_id"`
	Name            string    `json:"name"`
	CompletionDate  time.Time `json:"completion_date"`
	MinMembers      int32     `json:"min_members"`
	MaxMembers      int32     `json:"max_members"`
	ManagerID       string    `json:"manager_id"`
	ManagerUsername string    `json:"manager_username"`
	ManagerRole     string    `json:"manager_role"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type ProjectTaskCreateEvent struct {
	ProjectID  string    `json:"project_id"`
	Task       Task      `json:"task"`
	OccurredAt time.Time `json:"occurred_at"`
}

type ProjectAddMemberEvent struct {
	ProjectID       string    `json:"project_id"`
	Name            string    `json:"name"`
	CompletionDate  time.Time `json:"completion_date"`
	MinMembers      int32     `json:"min_members"`
	MaxMembers      int32     `json:"max_members"`
	ManagerID       string    `json:"manager_id"`
	ManagerUsername string    `json:"manager_username"`
	ManagerRole     string    `json:"manager_role"`
	MemberToAdd     User      `json:"member_to_add"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type ProjectRemoveMemberEvent struct {
	ProjectID        string    `json:"project_id"`
	Name             string    `json:"name"`
	CompletionDate   time.Time `json:"completion_date"`
	MinMembers       int32     `json:"min_members"`
	MaxMembers       int32     `json:"max_members"`
	ManagerID        string    `json:"manager_id"`
	ManagerUsername  string    `json:"manager_username"`
	ManagerRole      string    `json:"manager_role"`
	MemberToRemoveId string    `json:"member_to_remove"`
	OccurredAt       time.Time `json:"occurred_at"`
}

type ProjectTaskAddMemberEvent struct {
	ProjectID   string    `json:"project_id"`
	TaskName    string    `json:"task_name"`
	TaskId      string    `json:"task_id"`
	MemberToAdd User      `json:"member_to_add"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type ProjectTaskRemoveMemberEvent struct {
	ProjectID     string    `json:"project_id"`
	TaskName      string    `json:"task_name"`
	TaskId        string    `json:"task_id"`
	MemberToAddId string    `json:"member_to_add_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type ProjectTaskStatusEvent struct {
	ProjectID     string    `json:"project_id"`
	TaskId        string    `json:"task_id"`
	TaskStatus    string    `json:"task_status"`
	TaskNewStatus string    `json:"task_new_status"`
	Members       []User    `json:"members"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type TaskUpdateFileEvent struct {
	TaskId      string    `json:"task_id"`
	UserId      string    `json:"user_id"`
	FileName    string    `json:"file_name"`
	FileContent []byte    `json:"file_content"`
	OccurredAt  time.Time `json:"occurred_at"`
}
