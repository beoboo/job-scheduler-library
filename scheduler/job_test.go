package scheduler

import (
	"github.com/beoboo/job-scheduler/library/stream"
	"testing"
	"time"
)

func TestJobStart(t *testing.T) {
	j := newJob()

	assertJobStatus(t, j, Idle, -1)

	_ = j.start("sleep", "0.1")

	assertJobStatus(t, j, Running, -1)

	if j.id == "" {
		t.Fatalf("job PID should not be empty")
	}

	j.wait()

	assertJobStatus(t, j, Exited, 0)
}

func TestUnknownExecutable(t *testing.T) {
	j := newJob()

	_ = j.start("./unknown-executable")

	assertJobStatus(t, j, Errored, -1)
}

func TestJobStop(t *testing.T) {
	j := newJob()

	assertJobStatus(t, j, Idle, -1)

	_ = j.start("sleep", "1")

	assertJobStatus(t, j, Running, -1)

	if j.id == "" {
		t.Fatalf("job PID should not be empty")
	}

	_ = j.stop()
	assertJobStatus(t, j, Killed, -1)
}

func TestJobOutput(t *testing.T) {
	j := newJob()

	assertJobStatus(t, j, Idle, -1)

	expected := []string{
		"Running for 2 times, sleeping for 0.1\n",
		"#1\n",
		"#2\n",
	}

	err := j.start("../test.sh", "2", "0.1")
	if err != nil {
		t.Fatal(err)
	}

	assertJobOutput(t, j, expected)

	j.wait()

	assertJobStatus(t, j, Exited, 0)

	assertJobOutput(t, j, expected)
}

func TestJobMultipleReaders(t *testing.T) {
	j := newJob()

	expected := []string{
		"Running for 2 times, sleeping for 0.1\n",
		"#1\n",
		"#2\n",
	}

	err := j.start("../test.sh", "2", "0.1")
	if err != nil {
		t.Fatal(err)
	}

	o1 := j.output()

	time.Sleep(10 * time.Millisecond)

	o2 := j.output()

	j.wait()

	assertOutput(t, o1, expected)
	assertOutput(t, o2, expected)
}

func assertJobStatus(t *testing.T, j *job, expectedType StatusType, expectedCode int) {
	assertStatus(t, j.status(), expectedType, expectedCode)
}

func assertJobOutput(t *testing.T, j *job, expected []string) {
	o := j.output()

	assertOutput(t, o, expected)
}

func assertStatus(t *testing.T, st *JobStatus, expectedType StatusType, expectedCode int) {
	if st.Type != expectedType {
		t.Fatalf("job status should be \"%s\", got \"%s\"", expectedType, st.Type)
	}
	if st.ExitCode != expectedCode {
		t.Fatalf("job exit code should be \"%d\", got \"%d\"", expectedCode, st.ExitCode)
	}
}

func assertOutput(t *testing.T, o *stream.Stream, expected []string) {
	lines := o.Read()
	defer o.Unsubscribe()

	for _, e := range expected {
		line := <-lines

		if string(line.Text) != e {
			t.Fatalf("job output should contain \"%s\"%d, got \"%s\"%d", e, len(e), line.Text, len(line.Text))
		}
	}
}
