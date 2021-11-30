package status

import "fmt"

const (
	IDLE    = "idle"
	RUNNING = "running"
	EXITED  = "exited"
	KILLED  = "killed"
)

type Status struct {
	Value    string
	ExitCode int
}

func Idle() *Status {
	return create(IDLE, -1)
}

func Running() *Status {
	return create(RUNNING, -1)
}

func Exited(exitCode int) *Status {
	return create(EXITED, exitCode)
}

func Killed(exitCode int) *Status {
	return create(KILLED, exitCode)
}

func create(status string, exitCode int) *Status {
	return &Status{
		Value:    status,
		ExitCode: exitCode,
	}
}

func (s *Status) String() string {
	return fmt.Sprintf("%s (%d)", s.Value, s.ExitCode)
}

func (s *Status) Clone() *Status {
	return &Status{
		Value:    s.Value,
		ExitCode: s.ExitCode,
	}
}
