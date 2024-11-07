package domain

type Status int

const (
	Pending Status = 0
	Working Status = 1
	Done    Status = 2
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
