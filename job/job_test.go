package job

import (
	"github.com/beoboo/job-scheduler/library/stream"
	"testing"
	"time"
)

func TestJobStart(t *testing.T) {
	j := New()

	assertStatus(t, j, Idle, -1)

	_ = j.Start("sleep", "0.1")

	assertStatus(t, j, Running, -1)

	if j.Id() == "" {
		t.Fatalf("Job PID should not be empty")
	}

	j.Wait()

	assertStatus(t, j, Exited, 0)
}

func TestJobStop(t *testing.T) {
	j := New()

	assertStatus(t, j, Idle, -1)

	_ = j.Start("sleep", "1")

	assertStatus(t, j, Running, -1)

	if j.Id() == "" {
		t.Fatalf("Job PID should not be empty")
	}

	_ = j.Stop()
	assertStatus(t, j, Killed, -1)
}

func TestJobOutput(t *testing.T) {
	j := New()

	assertStatus(t, j, Idle, -1)

	expected := []string{
		"Running for 2 times, sleeping for 0.1\n",
		"#1\n",
		"#2\n",
	}

	err := j.Start("../test.sh", "2", "0.1")
	if err != nil {
		t.Fatal(err)
	}

	assertJobOutput(t, j, expected)

	j.Wait()

	assertStatus(t, j, Exited, 0)

	assertJobOutput(t, j, expected)
}

func TestJobMultipleReaders(t *testing.T) {
	j := New()

	expected := []string{
		"Running for 2 times, sleeping for 0.1\n",
		"#1\n",
		"#2\n",
	}

	err := j.Start("../test.sh", "2", "0.1")
	if err != nil {
		t.Fatal(err)
	}

	o1 := j.Output()

	time.Sleep(10 * time.Millisecond)

	o2 := j.Output()

	j.Wait()

	assertOutput(t, o1, expected)
	assertOutput(t, o2, expected)
}

func assertStatus(t *testing.T, j *Job, expectedType StatusType, expectedExitCode int) {
	st := j.Status()

	if st.Type != expectedType {
		t.Fatalf("Job status should be \"%s\", got \"%s\"", expectedType, st.Type)
	}
	if st.ExitCode != expectedExitCode {
		t.Fatalf("Job exit code should be \"%d\", got \"%d\"", expectedExitCode, st.ExitCode)
	}
}

func assertJobOutput(t *testing.T, j *Job, expected []string) {
	o := j.Output()

	assertOutput(t, o, expected)
}

func assertOutput(t *testing.T, o *stream.Stream, expected []string) {
	lines := o.Read()
	for _, e := range expected {
		line := <-lines

		if string(line.Text) != e {
			t.Fatalf("Job output should contain \"%s\"%d, got \"%s\"%d", e, len(e), line.Text, len(line.Text))
		}
	}
}
