package scheduler

import (
	"bufio"
	"fmt"
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/logsync"
	"github.com/beoboo/job-scheduler/library/stream"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
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
	sts      *JobStatus
	m        logsync.Mutex
	wg       *logsync.WaitGroup
}

// newJob creates a new job
func newJob(wg *logsync.WaitGroup) *job {
	id := generateRandomId()
	p := &job{
		id:       id,
		outputSt: stream.New(),
		sts:      &JobStatus{Type: Idle, ExitCode: -1},
		m:        logsync.NewMutex(fmt.Sprintf("job %s", id)),
		wg:       wg,
	}

	return p
}

func generateRandomId() string {
	return uuid.New().String()
}

// startIsolated starts the execution of a process through a parent/child mechanism
func (j *job) startIsolated(executable string, mem int, args ...string) error {
	log.Debugf("Starting isolated: %s\n", helpers.FormatCmdLine(executable, args...))
	cmd := exec.Command(executable, args...)

	// TODO: for simplicity, we're not handling other namespaces (i.e. UTS) or UID/GID mappings
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET,
	}

	j.cmd = cmd

	errCh := make(chan error, 1)

	j.wg.Add(j.id, 1)

	go func() {
		defer j.cleanupIsolated()

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

func (j *job) cleanupIsolated() {
	j.wg.Done(j.id)
}

// startChild starts the execution of a child process, capturing its output
func (j *job) startChild(jobId, executable string, mem int, args ...string) (int, error) {
	log.Debugf("Starting child [%s]: %s\n", jobId, helpers.FormatCmdLine(executable, args...))
	defer j.cleanupChild()

	// TODO: set cgroups
	// TODO: mount folders
	// TODO: chroot or pivot_root
	// TODO: cd /

	// TODO: set cgroups for CPU/IO
	if err := j.cgroups(jobId, mem); err != nil {
		return -1, err
	}

	cmd := exec.Command(executable, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return cmd.ProcessState.ExitCode(), err
	}

	return cmd.ProcessState.ExitCode(), nil
}

func (j *job) cgroups(jobId string, mem int) error {
	/*
		TODO: handle resources in a common/configurable base path so that it can be cleaned up easily
		i.e. /sys/fs/groups/FOO/[JOB_ID]
	*/
	if mem > 0 {
		log.Debugf("Setting memory limit for %s to %d\n", jobId, mem)
		dir := "/sys/fs/cgroup/memory/" + jobId

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory \"%s\": %v\n", dir, err)
		}

		if err := ioutil.WriteFile(dir+"/memory.limit_in_bytes", itob(mem), 0644); err != nil {
			return fmt.Errorf("unable to write memory limit: %v", err)
		}

		if err := ioutil.WriteFile(dir+"/cgroup.procs", itob(os.Getpid()), 0700); err != nil {
			return fmt.Errorf("unable to write to cgroup.procs file: %v", err)
		}
	}

	return nil
}

func itob(num int) []byte {
	return []byte(itoa(num))
}

func itoa(num int) string {
	return strconv.Itoa(num)
}

func (j *job) cleanupChild() {
	// TODO: unmount folders
	// TODO: cleanupChild cgroups
}

// stop stops a running process
func (j *job) stop() error {
	j.m.WLock("stop")
	if j.cmd == nil {
		j.m.WUnlock("stop")
		return fmt.Errorf("job not started")
	}
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

func (j *job) run(started chan error) error {
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

	// Wait is expected to fail for three reasons, none of which apply here:
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

	return nil
}

func (j *job) pipe(st stream.StreamType, pipe *bufio.Reader, wg *sync.WaitGroup) {
	for {
		buf := make([]byte, BUFFER_SIZE)
		n, err := pipe.Read(buf)

		if n > 0 {
			log.Debugf("[%s] %s", st, buf[:n])

			wErr := j.write(st, buf[:n])
			if wErr != nil {
				// The stream is already has been closed, due to the updated status of the job (that's
				// no more running). This means that the underlying execution has already finished, and
				// its pipes are already closed.
				log.Debugf("Write error: %s\n", wErr)
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
	log.Tracef("Updating status for job [%s] to %s\n", j.id, st)
	j.m.WLock("updateStatus")
	defer j.m.WUnlock("updateStatus")

	switch j.sts.Type {
	case Exited, Errored, Killed:
		// Do not update the status, the previous one is the one we want to keep
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
