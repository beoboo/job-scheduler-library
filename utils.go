package main

import (
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/status"
	"github.com/beoboo/job-scheduler/library/stream"
	"io"
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
	for {
		l, err := o.Read()
		if err == io.EOF {
			break
		}

		check(err)

		if l.Channel == stream.Output {
			log.Infoln(l)
		} else {
			log.Warnln(l)
		}
	}
}

func printStatus(status *status.Status, err error) {
	if err != nil {
		log.Fatalln(err)
	}
	log.Infof("Job status: %s\n", status)
}
