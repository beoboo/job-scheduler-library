//+build linux

package scheduler

import (
	"fmt"
	"github.com/beoboo/job-scheduler/library/stream"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func init() {
	os.Setenv("PATH", fmt.Sprintf("../bin:%s", os.Getenv("PATH")))
}

func TestIsolatedProcessCannotKillParent(t *testing.T) {
	checkDaemon(t)

	var s = New("worker")

	id, err := s.Start("kill", 0, strconv.Itoa(os.Getppid()))
	if err != nil {
		t.Fatalf("Job not started: %v\n", err)
	}

	o, err := s.Output(id)
	if err != nil {
		t.Fatalf("Cannot get job output: %v\n", err)
	}

	res := collect(o)

	expected := "No such process"
	if !strings.Contains(res, expected) {
		t.Fatalf("Expected \"%s\" to be in \"%s\"", expected, res)
	}

	s.Wait()
}

func TestIsolatedNetworkNamespace(t *testing.T) {
	checkDaemon(t)

	var s = New("worker")

	id, err := s.Start("route", 0)
	if err != nil {
		t.Fatalf("Job not started: %v\n", err)
	}

	s.Wait()

	st, _ := s.Status(id)

	// Would be 0 if the command execute successfully
	if st.ExitCode != 1 {
		t.Fatalf("Expected exit code: %d, got %d", -1, st.ExitCode)
	}
}

func TestMemoryLimit(t *testing.T) {
	checkDaemon(t)

	var s = New("worker")
	id1, err := s.Start("../bin/test.sh", 5000000, "1 10")
	if err != nil {
		t.Fatalf("Job not started: %v\n", err)
	}

	pid := s.jobs[id1].pid()
	id2, err := s.Start("ps", 0, "-o", "cgroup", strconv.Itoa(pid))
	if err != nil {
		t.Fatalf("Job not started: %v\n", err)
	}

	o, err := s.Output(id2)
	if err != nil {
		t.Fatalf("Cannot get job output: %v\n", err)
	}

	res := collect(o)

	expected := fmt.Sprintf("memory:/%s", id1)
	if !strings.Contains(res, expected) {
		t.Fatalf("Expected \"%s\" to be in \"%s\"", expected, res)
	}

	s.Wait()
}

func checkDaemon(t *testing.T) {
	_, err := exec.LookPath("../bin/worker")
	if err != nil {
		t.Skip(`No \"worker\" found in \"../bin\". Please build it with:

  go build -O bin ./cmd/worker
`)
	}
}

func collect(o *stream.Stream) string {
	res := ""
	done := make(chan bool)
	go func() {
		for l := range o.Read() {
			res += string(l.Text)
		}
		done <- true
	}()

	<-done
	return res
}
