package scheduler

import (
	"bufio"
	"fmt"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/logsync"
	"github.com/beoboo/job-scheduler/library/stream"
	"github.com/google/uuid"
	"os/exec"
	"sync"
	"time"
)

const (
	BUFFER_SIZE = 1024
)

// job wraps the execution of a process, capturing its stdout and stderr streams,
// and providing the process status
type job struct {
	id       string
	cmd      *exec.Cmd
	outputSt *stream.Stream
	done     chan bool
	sts      *JobStatus
	m        logsync.Mutex
}

// newJob creates a new job
func newJob() *job {
	id := generateRandomId()
	p := &job{
		id:       id,
		done:     make(chan bool),
		outputSt: stream.New(),
		sts:      &JobStatus{Type: Idle, ExitCode: -1},
		m:        logsync.New(fmt.Sprintf("job %s", id)),
	}

	return p
}

func generateRandomId() string {
	return uuid.New().String()
}

// startIsolated starts the execution of a process through a parent/child mechanism
func (j *job) startIsolated(executable string, args ...string) error {
	// TODO: This will run the process through /proc/self/exe (or similar). For now just wraps StartChild
	return j.start(executable, args...)
}

// start starts the execution of a child process, capturing its output
func (j *job) start(executable string, args ...string) error {
	cmd := exec.Command(executable, args...)
	// TODO: here we'll set clone flags
	// TODO: here we'll mounts and cgroups for the child process

	j.cmd = cmd

	errCh := make(chan error, 1)
	log.Debugf("Starting: %s\n", j.cmd.Path)

	go func() {
		err := j.run(errCh)
		if err != nil {
			j.updateStatus(Errored)
			errCh <- err
		}
	}()

	// Waits for the job to be started successfully
	err := <-errCh

	return err
}

// stop stops a running process
func (j *job) stop() error {
	j.m.WLock("stop")
	err := j.cmd.Process.Kill()
	j.m.WUnlock("stop")

	if err != nil {
		return fmt.Errorf("cannot kill job %d: (%s)", j.pid(), err)
	}

	j.updateStatus(Killed)
	return nil
}

// output returns the stream of captured stdout/stderr of the running process.
func (j *job) output() *stream.Stream {
	j.m.RLock("output")
	defer j.m.RUnlock("output")

	return j.outputSt
}

// status returns the current status of the job
func (j *job) status() *JobStatus {
	j.m.RLock("status")
	defer j.m.RUnlock("status")

	// JobStatus needs to be cloned or a race condition might happen (since it's a reference)
	return j.sts.clone()
}

// wait blocks until the process is completed
func (j *job) wait() {
	<-j.done
}

func (j *job) run(started chan error) error {
	log.Debugf("Running: %s\n", j.cmd.Path)

	stdout, err := j.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := j.cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	var wg sync.WaitGroup
	wg.Add(2)
	go j.pipe(stream.Output, stdoutReader, &wg)
	go j.pipe(stream.Error, stderrReader, &wg)

	err = j.cmd.Start()
	if err != nil {
		return err
	}

	j.updateStatus(Running)

	started <- err

	wg.Wait()
	// wait is expected to fail for three reasons, none of which apply here
	// 1. the process is not started (but that doesn't apply, or it would have exited above)
	// 2. it's called twice (not the case, again)
	// 3. the process is killed (and we can just log this and return the status and exit code)
	err = j.cmd.Wait()

	j.updateExitCode()

	if err != nil {
		log.Debugf("Error calling wait: %v\n", err)
		j.updateStatus(Errored)
	} else {
		j.updateStatus(Exited)
	}

	j.done <- true

	return nil
}

func (j *job) pipe(st stream.StreamType, pipe *bufio.Reader, wg *sync.WaitGroup) {
	for {
		buf := make([]byte, BUFFER_SIZE)
		n, err := pipe.Read(buf)

		if n > 0 {
			log.Debugf("[%s] %s", st, buf[:n])

			err := j.write(st, buf[:n])
			if err != nil {
				// The stream is already has been closed, due to the updated status of the job (that's
				// no more running). This means that the underlying execution has already finished, and
				// its pipes are already closed.
				log.Debugf("Write error: %s\n", err)
				break
			}
		}

		if err != nil {
			log.Debugf("Pipe closed: %s\n", err)
			break
		}
	}

	wg.Done()
}

func (j *job) write(st stream.StreamType, text []byte) error {
	j.m.WLock("pipe")
	defer j.m.WUnlock("pipe")

	return j.outputSt.Write(stream.Line{
		Time: time.Now(),
		Type: st,
		Text: text,
	})
}

func (j *job) pid() int {
	return j.cmd.Process.Pid
}

func (j *job) updateStatus(st StatusType) {
	j.m.WLock("updateStatus")
	defer j.m.WUnlock("updateStatus")

	switch j.sts.Type {
	case Running:
		j.sts.Type = st
		j.outputSt.Close()
	default:
		j.sts.Type = st
	}
}

func (j *job) updateExitCode() {
	j.m.WLock("updateExitCode")
	defer j.m.WUnlock("updateExitCode")

	j.sts.ExitCode = j.cmd.ProcessState.ExitCode()
}
