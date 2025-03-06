package models

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type Task struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      Status             `bson:"status" json:"status"`
	ProjectID   string             `bson:"project_id" json:"project_id"`
	Members     []User             `bson:"members" json:"members"`
}

type User struct {
	Id       string `bson:"_id,omitempty" json:"id"`
	Username string `bson:"username" json:"username"`
	Role     string `bson:"role" json:"role"`
}

type Status int

const (
	Pending Status = iota
	Working
	Done
)

func (s Status) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Working:
		return "Working"
	case Done:
		return "Done"
	default:
		return "Unknown"
	}
}

func ParseTaskStatus(status int) (Status, error) {
	switch status {
	case int(Pending):
		return Pending, nil
	case int(Working):
		return Working, nil
	case int(Done):
		return Done, nil
	default:
		return Pending, fmt.Errorf("invalid status value: %d", status)
	}
}
func ParseTaskStatus2(status string) (Status, error) {
	switch status {
	case "Pending":
		return Pending, nil
	case "Working":
		return Working, nil
	case "Done":
		return Done, nil
	default:
		return Pending, fmt.Errorf("invalid status value: %d", status)
	}
}
