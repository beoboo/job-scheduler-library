package scheduler

import (
	"github.com/beoboo/job-scheduler/library/assert"
	"github.com/beoboo/job-scheduler/library/status"
	"testing"
	"time"
)

var s = New(false)

func TestStartStop(t *testing.T) {
	id, _ := s.Start("echo", "world")

	if s.Size() != 1 {
		t.Fatalf("Job not started")
	}

	assertStatus(t, s, id, status.Running())

	time.Sleep(100 * time.Millisecond)

	assertStatus(t, s, id, status.Exited(0))
}

func TestStop(t *testing.T) {
	id, _ := s.Start("echo", "world")

	assertStatus(t, s, id, status.Running())

	_, _ = s.Stop(id)

	assertStatus(t, s, id, status.Killed(-1))
}

func TestOutput(t *testing.T) {
	expectedLines := []string{
		"Running for 1 times, sleeping for 0.1",
		"#1",
	}

	id, _ := s.Start("../test.sh", "1", "0.1")

	time.Sleep(150 * time.Millisecond)

	assertOutput(t, s, id, expectedLines)

	_, _ = s.Stop(id)

	assertOutput(t, s, id, expectedLines)
}

func assertStatus(t *testing.T, s *Scheduler, id string, expected *status.Status) {
	st, _ := s.Status(id)
	assert.AssertStatus(t, st, expected)
}

func assertOutput(t *testing.T, s *Scheduler, id string, expected []string) {
	o, err := s.Output(id)
	if err != nil {
		t.Fatalf("Expected output for job %s\n", id)
	}
	assert.AssertOutput(t, o, expected)
}
