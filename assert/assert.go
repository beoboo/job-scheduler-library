package assert

import (
	"github.com/beoboo/job-scheduler/library/status"
	"github.com/beoboo/job-scheduler/library/stream"
	"testing"
)

func AssertStatus(t *testing.T, st *status.Status, expected *status.Status) {
	if st.Value != expected.Value {
		t.Fatalf("Job status should be \"%s\", got \"%s\"", expected.Value, st.Value)
	}
	if st.ExitCode != expected.ExitCode {
		t.Fatalf("Job exit code should be \"%d\", got \"%d\"", expected.ExitCode, st.ExitCode)
	}
}

func AssertOutput(t *testing.T, o *stream.Stream, expected []string) {
	for _, e := range expected {
		l, err := o.Read()
		if err != nil {
			t.Fatal("Expected output line")
		}

		if string(l.Text) != e {
			t.Fatalf("Job output should contain \"%s\", got \"%s\"", e, l.Text)
		}
	}
}
