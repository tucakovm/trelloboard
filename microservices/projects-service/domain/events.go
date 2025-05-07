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
