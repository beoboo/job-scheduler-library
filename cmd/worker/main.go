package main

import (
	"flag"
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

	s := scheduler.New("worker")

	switch command {
	case "child":
		if len(args) < 2 {
			log.Fatalf("Usage: child [--cpu N] [--io N] [--mem N] JOB_ID EXECUTABLE [ARGS]\n")
		}

		// TODO: use a better arg/option parsing lib
		// TODO: handle cmd line options and limits for CPU/IO
		fs := flag.NewFlagSet("child", flag.ContinueOnError)
		mem := fs.Int("mem", 0, "Max memory usage in MB")
		err := fs.Parse(args)
		if err != nil {
			log.Fatalf("Cannot parse arguments: %s\n", err)
		}
		remaining := fs.Args()

		runChild(s, os.Args[0], *mem, remaining...)
	default:
		log.Fatalf(usage)
	}

	log.Reset()
}

func runChild(s *scheduler.Scheduler, executable string, mem int, params ...string) {
	log.Infof("Starting scheduler with \"%s\"\n", helpers.FormatCmdLine(executable, params...))
	_, err := s.Start(executable, mem, params...)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
		return
	}

	s.Wait()
}
