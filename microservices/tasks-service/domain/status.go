package domain

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
