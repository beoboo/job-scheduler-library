package scheduler

import (
	"fmt"
	"github.com/beoboo/job-scheduler/library/errors"
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/job"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/logsync"
	"github.com/beoboo/job-scheduler/library/stream"
)

type Scheduler struct {
	runJobAsChild bool // For debugging purposes only
	jobs          map[string]*job.Job
	m             logsync.Mutex
}

// New creates a scheduler.
func New(runJobAsChild bool) *Scheduler {
	return &Scheduler{
		runJobAsChild: runJobAsChild,
		jobs:          make(map[string]*job.Job),
		m:             logsync.New("Scheduler"),
	}
}

// Start runs a new job.Job.
func (s *Scheduler) Start(executable string, args ...string) (string, error) {
	log.Debugf("Starting executable: \"%s\"\n", helpers.FormatCmdLine(executable, args...))
	j := job.New()

	// For debugging purposes, bypasses the creation of the spawned child
	if s.runJobAsChild {
		err := j.StartChild(executable, args...)
		if err != nil {
			return "", err
		}
	} else {
		err := j.Start(executable, args...)
		if err != nil {
			return "", err
		}
	}

	log.Debugf("Job ID: %s\n", j.Id())
	log.Debugf("Status: %s\n", j.Status())

	s.m.WLock("Start")
	defer s.m.WUnlock("Start")

	s.jobs[j.Id()] = j
	return j.Id(), nil
}

// Stop stops a running job.Job, or an error if the job.Job doesn't exist.
func (s *Scheduler) Stop(id string) (*job.Status, error) {
	log.Debugf("Stopping job %s\n", id)

	s.m.RLock("Stop")
	defer s.m.RUnlock("Stop")
	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	err := j.Stop()
	if err != nil {
		return nil, fmt.Errorf("cannot stop job: %s", id)
	}

	return j.Status(), nil
}

// Status returns the status of a job.Job, or an error if the job.Job doesn't exist.
func (s *Scheduler) Status(id string) (*job.Status, error) {
	log.Debugf("Checking status for job \"%s\"\n", id)

	s.m.RLock("Status")
	defer s.m.RUnlock("Status")

	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	return j.Status(), nil
}

// Output returns the stream of the stdout/stderr of a job.Job, or an error if the job.Job doesn't exist.
func (s *Scheduler) Output(id string) (*stream.Stream, error) {
	log.Debugf("Streaming output for job \"%s\"\n", id)

	s.m.RLock("Output")
	defer s.m.RUnlock("Output")
	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	return j.Output(), nil
}

// Size returns the number of stored jobs.
func (s *Scheduler) Size() int {
	s.m.RLock("Size")
	defer s.m.RUnlock("Size")

	return len(s.jobs)
}
