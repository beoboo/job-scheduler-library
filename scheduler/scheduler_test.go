package scheduler

import (
	"github.com/beoboo/job-scheduler/library/job"
	"testing"
	"time"
)

var s = New(false)

func TestSchedulerStart(t *testing.T) {
	id, _ := s.Start("sleep", "0.1")

	if s.Size() != 1 {
		t.Fatalf("Job not started")
	}

	assertStatus(t, s, id, job.Running, -1)

	time.Sleep(200 * time.Millisecond)

	assertStatus(t, s, id, job.Exited, 0)
}

func TestSchedulerStop(t *testing.T) {
	id, _ := s.Start("sleep", "10")

	assertStatus(t, s, id, job.Running, -1)

	_, _ = s.Stop(id)

	assertStatus(t, s, id, job.Killed, -1)
}

func TestSchedulerOutput(t *testing.T) {
	expectedLines := []string{
		"Running for 1 times, sleeping for 0.1\n",
		"#1\n",
	}

	id, _ := s.Start("../test.sh", "1", "0.1")

	time.Sleep(150 * time.Millisecond)

	assertOutput(t, s, id, expectedLines)

	_, _ = s.Stop(id)

	assertOutput(t, s, id, expectedLines)
}

func assertStatus(t *testing.T, s *Scheduler, id string, expectedType job.StatusType, expectedExitCode int) {
	st, _ := s.Status(id)
	if st.Type != expectedType {
		t.Fatalf("Job status should be \"%s\", got \"%s\"", expectedType, st.Type)
	}
	if st.ExitCode != expectedExitCode {
		t.Fatalf("Job exit code should be \"%d\", got \"%d\"", expectedExitCode, st.ExitCode)
	}
}

func assertOutput(t *testing.T, s *Scheduler, id string, expected []string) {
	o, err := s.Output(id)
	if err != nil {
		t.Fatalf("Expected output for job %s\n", id)
	}

	lines := o.Read()
	defer o.Unsubscribe()
	for _, e := range expected {
		line := <-lines

		if string(line.Text) != e {
			t.Fatalf("Job output should contain \"%s\"%d, got \"%s\"%d", e, len(e), line.Text, len(line.Text))
		}
	}
}
