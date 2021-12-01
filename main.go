package main

import (
	"flag"
	"github.com/beoboo/job-scheduler/library/log"
	"github.com/beoboo/job-scheduler/library/scheduler"
	"github.com/beoboo/job-scheduler/library/stream"
	"os"
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

	switch command {
	case "examples":
		// this will run some concurrent examples
	case "run":
		run(args...)
	case "child":
		// this will run a single process (without any job in the middle)
	default:
		log.Fatalf(usage)
	}
}

func isRoot() bool {
	return os.Geteuid() == 0
}

// Runs a job through the scheduler
func run(args ...string) {
	if len(args) < 1 {
		log.Fatalf("Usage: schedule [--cpu N] [--io N] [--mem N] EXECUTABLE [ARGS]\n")
	}

	err := flag.CommandLine.Parse(args)
	if err != nil {
		log.Fatalf("Cannot parse arguments: %s\n", err)
	}
	remaining := flag.Args()

	executable := remaining[0]
	params := remaining[1:]

	runScheduler(executable, params...)
}

func runScheduler(executable string, params ...string) {
	// TODO: map "dummy" to the runner (i.e. /proc/self/exe)
	s := scheduler.New("dummy")

	id := do(s.Start(executable, params...))
	log.Infof("job \"%s\" started\n", id)

	printStatus(s.Status(id))

	o, err := s.Output(id)
	if err != nil {
		log.Fatalf("Cannot retrieve job: %s\n", id)
	}

	printOutput(o)
	printStatus(s.Status(id))
}

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

		if line.Type == stream.Output {
			log.Infof("%s", line)
		} else {
			log.Warnf("%s", line)
		}
	}
}

func printStatus(status *scheduler.JobStatus, err error) {
	if err != nil {
		log.Fatalln(err)
	}
	log.Infof("job status: %s\n", status)
}
