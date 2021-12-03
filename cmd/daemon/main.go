package main

import (
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/scheduler"
	"os"
)

func main() {
	usage := "Usage: child\n"

	if len(os.Args) < 2 {
		log.Fatalf(usage)
	}

	command := os.Args[1]
	args := os.Args[2:]

	s := scheduler.New("daemon")

	switch command {
	case "child":
		log.SetMode(log.Dimmed)
		log.SetLevel(log.Info)

		if len(args) < 2 {
			log.Fatalf("Usage: child [--cpu N] [--io N] [--mem N] JOB_ID EXECUTABLE [ARGS]\n")
		}

		// TODO: handle cmd line options and limits
		//err := flag.CommandLine.Parse(args)
		//if err != nil {
		//	log.Fatalf("Cannot parse arguments: %s\n", err)
		//}
		//remaining := flag.Args()
		remaining := args

		runChild(s, os.Args[0], remaining...)
	default:
		log.Fatalf(usage)
	}

	log.Reset()
}

func runChild(s *scheduler.Scheduler, executable string, params ...string) {
	log.Infof("Starting scheduler with \"%s\"\n", helpers.FormatCmdLine(executable, params...))
	_, err := s.Start(executable, params...)
	if err != nil {
		log.Fatalf("Daemon error: %s\n", err)
		return
	}

	s.Wait()
}
