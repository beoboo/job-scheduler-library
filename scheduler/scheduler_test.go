package scheduler

import (
	"testing"
	"time"
)

var s = New("dummy")

func TestSchedulerStart(t *testing.T) {
	id, _ := s.Start("sleep", "0.1")

	if s.Size() != 1 {
		t.Fatalf("job not started")
	}

	assertSchedulerStatus(t, s, id, Running, -1)

	time.Sleep(200 * time.Millisecond)

	assertSchedulerStatus(t, s, id, Exited, 0)
}

func TestSchedulerStop(t *testing.T) {
	id, _ := s.Start("sleep", "10")

	assertSchedulerStatus(t, s, id, Running, -1)

	_, _ = s.Stop(id)

	assertSchedulerStatus(t, s, id, Killed, -1)
}

func TestSchedulerOutput(t *testing.T) {
	expected := []string{
		"Running for 1 times, sleeping for 0.1\n",
		"#1\n",
	}

	id, _ := s.Start("../test.sh", "1", "0.1")

	time.Sleep(150 * time.Millisecond)

	assertSchedulerOutput(t, s, id, expected)

	_, _ = s.Stop(id)

	assertSchedulerOutput(t, s, id, expected)
}

func assertSchedulerStatus(t *testing.T, s *Scheduler, id string, expectedStatusType StatusType, expectedExitCode int) {
	st, _ := s.Status(id)
	assertStatus(t, st, expectedStatusType, expectedExitCode)
}

func assertSchedulerOutput(t *testing.T, s *Scheduler, id string, expected []string) {
	o, _ := s.Output(id)
	assertOutput(t, o, expected)
}
