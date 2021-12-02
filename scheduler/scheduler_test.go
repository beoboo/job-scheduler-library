package scheduler

import (
	"strings"
	"testing"
	"time"
)

const Runner = "../scripts/echo.sh"

var s = New(Runner)

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
		Runner,
	}

	id, _ := s.Start("../scripts/test.sh", "1", "0.1")

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

	lines := o.Read()

	for _, e := range expected {
		line := <-lines

		if !strings.Contains(string(line.Text), e) {
			t.Fatalf("Job output should contain \"%s\", got \"%s\"", e, line.Text)
		}
	}
}
