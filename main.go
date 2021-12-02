package main

import (
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/scheduler"
	"github.com/beoboo/job-scheduler/library/stream"
	"os"
	"syscall"
)

func main() {
	usage := "Usage: examples|run|child\n"

	if !isRoot() {
		log.Fatalln("Please run this with root privileges.")
	}

	if len(os.Args) < 2 {
		log.Fatalf(usage)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// TODO: this could be set through an option
	log.SetLevel(log.Debug)

	//s := scheduler.New("scripts/echo.sh")
	s := scheduler.NewSelf()

	switch command {
	case "examples":
		runExamples(s)
	case "run":
		if len(args) < 1 {
			log.Fatalf("Usage: run [--cpu N] [--io N] [--mem N] EXECUTABLE [ARGS]\n")
		}

		// TODO: handle cmd line options and limits
		//err := flag.CommandLine.Parse(args)
		//if err != nil {
		//	log.Fatalf("Cannot parse arguments: %s\n", err)
		//}
		//remaining := flag.Args()
		remaining := args

		executable := remaining[0]
		params := remaining[1:]
		runParent(s, executable, params...)
	case "child":
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

func isRoot() bool {
	return os.Geteuid() == 0
}

func runParent(s *scheduler.Scheduler, executable string, params ...string) {
	log.Infof("Starting scheduler with \"%s\"\n", helpers.FormatCmdLine(executable, params...))
	id := do(s.Start(executable, params...))
	log.Infof("Job \"%s\" started\n", id)

	printStatus(s.Status(id))

	o, err := s.Output(id)
	if err != nil {
		log.Fatalf("Cannot retrieve job: %s\n", id)
	}

	printOutput(o)
	printStatus(s.Status(id))

	s.Wait()
	log.Infoln("Schedule completed")
}

func runChild(s *scheduler.Scheduler, executable string, params ...string) {
	log.Infof("Starting scheduler with \"%s\"\n", helpers.FormatCmdLine(executable, params...))
	_, err := s.Start(executable, params...)
	if err != nil {
		log.Errorln(err)
		syscall.Exit(127)
	}

	s.Wait()
}

func do(val string, err error) string {
	check(err)

	return val
}

func check(err error) {
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}
}

func printOutput(o *stream.Stream) {
	s := o.Read()

	for {
		line := <-s
		if line == nil {
			break
		}

		if line.Type == stream.Output {
			log.Infof("%s", line.Text)
		} else {
			log.Warnf("%s", line.Text)
		}
	}
}

func printStatus(status *scheduler.JobStatus, err error) {
	if err != nil {
		log.Fatalln(err)
	}
	log.Debugf("Job status: %s\n", status)
}
