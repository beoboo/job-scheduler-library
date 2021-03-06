package scheduler

import (
	"fmt"
	"github.com/beoboo/job-scheduler/library/errors"
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/logsync"
	"github.com/beoboo/job-scheduler/library/stream"
	"os"
)

const (
	Self = "/proc/self/exe"
)

type Scheduler struct {
	runner string
	jobs   map[string]*job
	m      logsync.Mutex
	wg     logsync.WaitGroup
}

func isRoot() bool {
	return os.Geteuid() == 0
}

// New creates a scheduler.
func New(runner string) *Scheduler {
	if !isRoot() {
		log.Fatalln("Please run this with root privileges.")
	}

	return &Scheduler{
		runner: runner,
		jobs:   make(map[string]*job),
		m:      logsync.NewMutex("Scheduler"),
		wg:     logsync.NewWaitGroup("Scheduler"),
	}
}

// NewSelf creates a scheduler for "/proc/self/exe".
func NewSelf() *Scheduler {
	return New(Self)
}

// Start runs a new job.
func (s *Scheduler) Start(executable string, mem int, args ...string) (string, error) {
	log.Debugf("Starting executable: \"%s\"\n", helpers.FormatCmdLine(executable, args...))
	j := newJob(&s.wg)

	// If the executable is not the same as the predefined runner, the process has to be isolated
	/**
		TODO: this is super simplified. We are checking that the name configured in the Scheduler
	 	is the same of the executable we're running (through the parent/child execution).
	    It works well with "/proc/self/exe", less for a "worker" or "child" binary that needs to be
	    under $PATH.
	*/
	if s.runner != executable {
		log.Debugln("Starting in isolated mode")
		args = append([]string{
			"child", // Main subcommand
			"--mem", itoa(mem),
			j.id,       // The job ID
			executable, // The original executable
		}, args...)

		err := j.startIsolated(s.runner, mem, args...)
		if err != nil {
			return "", err
		}

		log.Debugf("Job ID: %s\n", j.id)
		log.Debugf("Status: %s\n", j.status())

		s.m.WLock("Start")
		defer s.m.WUnlock("Start")

		s.jobs[j.id] = j

		return j.id, nil
	}

	jobId := args[0]
	executable = args[1]
	args = args[2:]

	log.Debugln("Starting in standard mode")

	ec, err := j.startChild(jobId, executable, mem, args...)
	if err != nil {
		log.Errorln(err)
	}

	os.Exit(ec)

	return "", nil
}

// Stop stops a running job, or an error if the job.job doesn't exist.
func (s *Scheduler) Stop(id string) (*JobStatus, error) {
	log.Debugf("Stopping job %s\n", id)

	s.m.RLock("stop")
	defer s.m.RUnlock("stop")
	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	err := j.stop()
	if err != nil {
		return nil, fmt.Errorf("cannot stop job: %s", id)
	}

	return j.status(), nil
}

// Status returns the status of a job, or an error if the job.job doesn't exist.
func (s *Scheduler) Status(id string) (*JobStatus, error) {
	log.Debugf("Checking status for job \"%s\"\n", id)

	s.m.RLock("Status")
	defer s.m.RUnlock("Status")

	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	return j.status(), nil
}

// Output returns the stream of the stdout/stderr of a job, or an error if the job.job doesn't exist.
func (s *Scheduler) Output(id string) (*stream.Stream, error) {
	//log.Debugf("Streaming output for job \"%s\"\n", id)

	s.m.RLock("Output")
	defer s.m.RUnlock("Output")
	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	return j.output(), nil
}

// Size returns the number of stored jobs.
func (s *Scheduler) Size() int {
	s.m.RLock("Size")
	defer s.m.RUnlock("Size")

	return len(s.jobs)
}

// Wait blocks until all jobs are finished
func (s *Scheduler) Wait() {
	s.wg.Wait()
}
