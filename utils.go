package main

import (
	"github.com/beoboo/job-scheduler/library/job"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/stream"
)

func do(val string, err error) string {
	check(err)

	return val
}

func check(err error) {
	if err != nil {
		log.Fatalf("Unexpected: %s\n", err)
	}
}

func printOutput(o *stream.Stream) {
	s := o.Read()

	for {
		line := <-s
		if line == nil {
			break
		}

		if line.Channel == stream.Output {
			log.Infoln(line)
		} else {
			log.Warnln(line)
		}
	}
}

func printStatus(status *job.Status, err error) {
	if err != nil {
		log.Fatalln(err)
	}
	log.Infof("Job status: %s\n", status)
}
