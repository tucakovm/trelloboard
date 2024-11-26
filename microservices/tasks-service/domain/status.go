package domain

import (
	"fmt"
)

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
