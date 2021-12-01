package job

import "fmt"

type StatusType int

const (
	Idle    StatusType = 0
	Running StatusType = 1
	Exited  StatusType = 2
	Killed  StatusType = 3
	Errored StatusType = 4
)

type Status struct {
	Type     StatusType
	ExitCode int
}

func (st StatusType) String() string {
	switch st {
	case Idle:
		return "idle"
	case Running:
		return "running"
	case Exited:
		return "exited"
	case Killed:
		return "killed"
	default:
		return "errored"
	}
}

func (s *Status) String() string {
	switch s.Type {
	case Idle | Running:
		return s.Type.String()
	default:
		return fmt.Sprintf("%s (%d)", s.Type, s.ExitCode)
	}
}

func (s *Status) Clone() *Status {
	return &Status{
		Type:     s.Type,
		ExitCode: s.ExitCode,
	}
}
