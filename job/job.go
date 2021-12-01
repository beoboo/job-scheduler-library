package job

import (
	"fmt"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/status"
	"github.com/beoboo/job-scheduler/library/stream"
	"github.com/google/uuid"
	"io"
	"os/exec"
	"sync"
	"time"
)

const (
	BUFFER_SIZE = 1024
)

// Job wraps the execution of a process, capturing its stdout and stderr streams,
// and providing the process status
type Job struct {
	id     string
	cmd    *exec.Cmd
	output *stream.Stream
	done   chan bool
	status *status.Status
	m      sync.RWMutex
}

// New creates a new Job
func New() *Job {
	p := &Job{
		id:     generateRandomId(),
		done:   make(chan bool),
		output: stream.New(),
		status: status.Idle(),
	}

	return p
}

func generateRandomId() string {
	return uuid.New().String()
}

// Id returns the job ID
func (j *Job) Id() string {
	return j.id
}

// Start starts the execution of a process through a parent/child mechanism
func (j *Job) Start(executable string, args ...string) error {
	// TODO: This will run the process through /proc/self/exe (or similar). For now just wraps StartChild
	return j.StartChild(executable, args...)
}

// StartChild starts the execution of a child process, capturing its output
func (j *Job) StartChild(executable string, args ...string) error {
	cmd := exec.Command(executable, args...)
	// TODO: here we'll set clone flags
	// TODO: here we'll mounts and cgroups for the child process

	j.cmd = cmd

	errCh := make(chan error, 1)
	log.Debugf("Starting: %s\n", j.cmd.Path)

	go func() {
		j.run(errCh)
	}()

	// Waits for the job to be started successfully
	err := <-errCh

	return err
}

// Stop stops a running process
func (j *Job) Stop() error {
	j.wlock("Stop")
	err := j.cmd.Process.Kill()
	j.wunlock("Stop")

	if err != nil {
		return fmt.Errorf("cannot kill job %d: (%s)", j.pid(), err)
	}

	j.updateStatus(status.KILLED)
	return nil
}

// Output returns the stream of captured stdout/stderr of the running process.
func (j *Job) Output() *stream.Stream {
	j.wlock("Output")
	defer j.wunlock("Output")

	return j.output
}

// Status returns the current status of the Job
func (j *Job) Status() *status.Status {
	j.rlock("Status")
	defer j.runlock("Status")

	// Status needs to be cloned or a race condition might happen (since it's a reference)
	return j.status.Clone()
}

// Wait blocks until the process is completed
func (j *Job) Wait() {
	<-j.done
}

func (j *Job) run(started chan error) {
	log.Debugf("Running: %s\n", j.cmd.Path)

	stdout, _ := j.cmd.StdoutPipe()
	stderr, _ := j.cmd.StderrPipe()

	err := j.cmd.Start()

	started <- err
	if err != nil {
		return
	}

	j.updateStatus(status.RUNNING)

	j.pipe(stream.Output, stdout)
	j.pipe(stream.Error, stderr)

	// Wait is expected to fail for three reasons
	// 1. the process is not started (but that doesn't apply, or it would have exited above)
	// 2. it's called twice (not the case, again)
	// 3. the process is killed (and we can just log this and return the status and exit code)
	err = j.cmd.Wait()
	j.updateExitCode()
	if err != nil {
		log.Debugf("Error calling Wait: %v\n", err)
	} else {
		j.updateStatus(status.EXITED)
	}

	j.done <- true
}

func (j *Job) pipe(channel stream.Channel, pipe io.ReadCloser) {
	go func() {
		for {
			buf := make([]byte, BUFFER_SIZE)
			n, err := pipe.Read(buf)
			if err != nil {
				log.Debugf("Pipe closed: %s\n", err)
				break
			}

			if n > 0 {
				log.Debugf("[%s] %s", channel, buf[:n])

				err := j.write(channel, buf[:n])
				if err != nil {
					break
				}
			} else {
				break
			}
		}
	}()
}

func (j *Job) write(channel stream.Channel, text []byte) error {
	j.wlock("pipe")
	defer j.wunlock("pipe")

	return j.output.Write(stream.Line{
		Channel: channel,
		Time:    time.Now(),
		Text:    text,
	})

}

func (j *Job) pid() int {
	return j.cmd.Process.Pid
}

func (j *Job) updateStatus(st string) {
	j.wlock("updateStatus")
	defer j.wunlock("updateStatus")

	// TODO: we could have a state machine here, and check for invalid state transitions
	switch j.status.Value {
	case status.IDLE:
		if st == status.RUNNING {
			j.status.Value = st
		}
	case status.RUNNING:
		if st == status.EXITED || st == status.KILLED {
			j.status.Value = st
			j.output.Close()
		}
	default:
		return
	}
}

func (j *Job) updateExitCode() {
	j.wlock("updateExitCode")
	defer j.wunlock("updateExitCode")

	// TODO: we could have a state machine here
	j.status.ExitCode = j.cmd.ProcessState.ExitCode()
}

// These wraps mutex R/W lock and unlock for debugging purposes
func (j *Job) rlock(id string) {
	log.Tracef("Job read locking %s\n", id)
	j.m.RLock()
}

func (j *Job) runlock(id string) {
	log.Tracef("Job read unlocking %s\n", id)
	j.m.RUnlock()
}

func (j *Job) wlock(id string) {
	log.Tracef("Job write locking %s\n", id)
	j.m.Lock()
}

func (j *Job) wunlock(id string) {
	log.Tracef("Job write unlocking %s\n", id)
	j.m.Unlock()
}
