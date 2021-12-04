package main

import (
	"flag"
	"github.com/beoboo/job-scheduler/library/helpers"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/scheduler"
	"github.com/beoboo/job-scheduler/library/stream"
	"os"
)

func main() {
	usage := "Usage: examples|run|child\n"

	if len(os.Args) < 2 {
		log.Fatalf(usage)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// TODO: this could be set through an option
	//log.SetLevel(log.Debug)

	// TODO: this could be set through an option
	//s := scheduler.New("scripts/echo.sh")
	s := scheduler.NewSelf()

	switch command {
	case "examples":
		runExamples(s)
	case "run":
		if len(args) < 1 {
			log.Fatalf("Usage: run [--cpu N] [--io N] [--mem N] EXECUTABLE [ARGS]\n")
		}

		mem, remaining := parseArgs(args)

		executable := remaining[0]
		args = remaining[1:]
		runParent(s, executable, mem, args...)
	case "child":
		if len(args) < 2 {
			log.Fatalf("Usage: child [--cpu N] [--io N] [--mem N] JOB_ID EXECUTABLE [ARGS]\n")
		}

		mem, remaining := parseArgs(args)

		runChild(s, os.Args[0], mem, remaining...)
	default:
		log.Fatalf(usage)
	}

	log.Reset()
}

func parseArgs(args []string) (int, []string) {
	// TODO: use a better arg/option parsing lib
	// TODO: handle cmd line options and limits for CPU/IO
	fs := flag.NewFlagSet("child", flag.ContinueOnError)
	mem := fs.Int("mem", 0, "Max memory usage in MB")
	err := fs.Parse(args)
	if err != nil {
		log.Fatalf("Cannot parseArgs arguments: %s\n", err)
	}
	remaining := fs.Args()

	return *mem, remaining
}

func runParent(s *scheduler.Scheduler, executable string, mem int, params ...string) {
	log.Infof("Starting scheduler with \"%s\"\n", helpers.FormatCmdLine(executable, params...))
	id := do(s.Start(executable, mem, params...))
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

func runChild(s *scheduler.Scheduler, executable string, mem int, params ...string) {
	log.Infof("Starting scheduler with \"%s\"\n", helpers.FormatCmdLine(executable, params...))
	_, err := s.Start(executable, mem, params...)
	if err != nil {
		log.Fatalf("Error: %v", err)
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
