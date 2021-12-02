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

	var s = New("daemon")

	id, err := s.Start("kill", strconv.Itoa(os.Getppid()))
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

func checkDaemon(t *testing.T) {
	_, err := exec.LookPath("../bin/daemon")
	if err != nil {
		t.Skip(`No \"daemon\" found in \"../bin\". Please build it with:

  go build -O bin ./cmd/daemon
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
