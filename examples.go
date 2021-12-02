package main

import (
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/scheduler"
	"sync"
)

func runExamples(s *scheduler.Scheduler) {
	var wg sync.WaitGroup

	// These examples will run concurrently, simulating several processes running at the same time.
	//runExample(1, example1, sched, &wg)
	runExample(2, example2NoExecutable, s, &wg)

	for i := 0; i < 10; i++ {
		//runExample(i, example1, sched, &wg)
	}

	wg.Wait()
}

func runExample(id int, example func(s *scheduler.Scheduler), s *scheduler.Scheduler, wg *sync.WaitGroup) {
	log.Infof("Example #%d\n", id)

	wg.Add(1)

	go func() {
		defer wg.Done()
		example(s)
	}()
}

func example1(s *scheduler.Scheduler) {
	id := do(s.Start("scripts/test.sh", "5", "1"))
	log.Infof("Job \"%s\" started\n", id)

	printStatus(s.Status(id))

	o, err := s.Output(id)
	if err != nil {
		log.Fatalf("Cannot retrieve job: %s\n", id)
	}

	for l := range o.Read() {
		log.Infof("%s", l)
	}

	printStatus(s.Status(id))
}

func example2NoExecutable(s *scheduler.Scheduler) {
	_, err := s.Start("./unknown")
	log.Warnf("Expected error: %s\n", err)
}
