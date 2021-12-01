package scheduler

import (
	"fmt"
	"github.com/beoboo/job-scheduler/library/errors"
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/job"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/stream"
	"sync"
)

type Scheduler struct {
	runJobAsChild bool // For debugging purposes only
	jobs          map[string]*job.Job
	m             sync.RWMutex
}

// New creates a scheduler.
func New(runJobAsChild bool) *Scheduler {
	return &Scheduler{
		runJobAsChild: runJobAsChild,
		jobs:          make(map[string]*job.Job),
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

	s.wlock("Start")
	defer s.wunlock("Start")

	s.jobs[j.Id()] = j
	return j.Id(), nil
}

// Stop stops a running job.Job, or an error if the job.Job doesn't exist.
func (s *Scheduler) Stop(id string) (*job.Status, error) {
	log.Debugf("Stopping job %s\n", id)

	s.rlock("Stop")
	defer s.runlock("Stop")
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

	s.rlock("Status")
	defer s.runlock("Status")

	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	return j.Status(), nil
}

// Output returns the stream of the stdout/stderr of a job.Job, or an error if the job.Job doesn't exist.
func (s *Scheduler) Output(id string) (*stream.Stream, error) {
	log.Debugf("Streaming output for job \"%s\"\n", id)

	s.rlock("Output")
	defer s.runlock("Output")
	j, ok := s.jobs[id]

	if !ok {
		return nil, &errors.NotFoundError{Id: id}
	}

	return j.Output(), nil
}

// Size returns the number of stored jobs.
func (s *Scheduler) Size() int {
	s.rlock("Size")
	defer s.runlock("Size")

	return len(s.jobs)
}

// These wraps mutex R/W lock and unlock for debugging purposes
func (s *Scheduler) rlock(id string) {
	log.Tracef("Scheduler read locking %s\n", id)
	s.m.RLock()
}

func (s *Scheduler) runlock(id string) {
	log.Tracef("Scheduler read unlocking %s\n", id)
	s.m.RUnlock()
}

func (s *Scheduler) wlock(id string) {
	log.Tracef("Scheduler write locking %s\n", id)
	s.m.Lock()
}

func (s *Scheduler) wunlock(id string) {
	log.Tracef("Scheduler write unlocking %s\n", id)
	s.m.Unlock()
}
