package job

import (
	"fmt"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/logsync"
	"github.com/beoboo/job-scheduler/library/stream"
	"github.com/google/uuid"
	"io"
	"os/exec"
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
	status *Status
	m      logsync.Mutex
}

// New creates a new Job
func New() *Job {
	id := generateRandomId()
	p := &Job{
		id:     id,
		done:   make(chan bool),
		output: stream.New(),
		status: &Status{Type: Idle, ExitCode: -1},
		m:      logsync.New(fmt.Sprintf("Job %s", id)),
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
	j.m.WLock("Stop")
	err := j.cmd.Process.Kill()
	j.m.WUnlock("Stop")

	if err != nil {
		return fmt.Errorf("cannot kill job %d: (%s)", j.pid(), err)
	}

	j.updateStatus(Killed)
	return nil
}

// Output returns the stream of captured stdout/stderr of the running process.
func (j *Job) Output() *stream.Stream {
	j.m.RLock("Output")
	defer j.m.RUnlock("Output")

	return j.output
}

// Status returns the current status of the Job
func (j *Job) Status() *Status {
	j.m.RLock("Status")
	defer j.m.RUnlock("Status")

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

	if err != nil {
		j.updateStatus(Errored)
		started <- err

		return
	}

	started <- err

	j.updateStatus(Running)

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
		j.updateStatus(Errored)
	} else {
		j.updateStatus(Exited)
	}

	j.done <- true
}

func (j *Job) pipe(channel stream.Channel, pipe io.ReadCloser) {
	go func() {
		for {
			buf := make([]byte, BUFFER_SIZE)
			n, err := pipe.Read(buf)

			if n > 0 {
				log.Debugf("[%s] %s", channel, buf[:n])

				err := j.write(channel, buf[:n])
				if err != nil {
					// The stream is already has been closed, due to the updated status of the job (that's
					// no more running). This means that the underlying execution has already finished, and
					// its pipes are already closed.
					break
				}
			}

			if err != nil {
				log.Debugf("Pipe closed: %s\n", err)
				break
			}
		}
	}()
}

func (j *Job) write(channel stream.Channel, text []byte) error {
	j.m.WLock("pipe")
	defer j.m.WUnlock("pipe")

	return j.output.Write(stream.Line{
		Channel: channel,
		Time:    time.Now(),
		Text:    text,
	})
}

func (j *Job) pid() int {
	return j.cmd.Process.Pid
}

func (j *Job) updateStatus(st StatusType) {
	j.m.WLock("updateStatus")
	defer j.m.WUnlock("updateStatus")

	switch j.status.Type {
	case Running:
		j.status.Type = st
		j.output.Close()
	default:
		j.status.Type = st
	}
}

func (j *Job) updateExitCode() {
	j.m.WLock("updateExitCode")
	defer j.m.WUnlock("updateExitCode")

	j.status.ExitCode = j.cmd.ProcessState.ExitCode()
}
