# Job Scheduler Library

The library exposes methods to:

* create a new job scheduler (in two different ways)
* start a job
* stop a job by its ID
* get the output of a job
* get the status
* wait for all jobs completion (this would be used only in static apps - like the example main - not in the server
  that’s already waiting for some other events).

There are two example applications:

* a [worker](cmd/worker/main.go), that's an example of how the library could be used to spin up child processes;
* a [main](main.go) app, that's showing:
    * concurrent example jobs;
    * how to run the scheduler (so, again, it starts a job, prints the output, prints the status, waits for completion,
      in a similar way as a server app);
    * how to run child jobs (in the same way as the worker does).

## Requirements

To test and run the project you need to have an environment configured with cgroup v1. The included Dockerfile is built
upon Ubuntu 20.04 that can simplify this on Mac machines running Docker Desktop.

To spin that up, run

```
make docker-build
make docker-run
```

and then check that you have the required cgroup version with

```
mount | grep cgroup
```

Another possible setup is to run Virtualbox with the same Ubuntu 20.04 running on it.

## Building the worker

The project contains an [example](cmd/worker/main.go) of an executable that can answer to "child" commands.

It can be built with:

```
go build -o bin/ ./cmd/worker
```

_Note: Some tests assume that this application is contained the `bin` folder._

## Testing

Run the tests with

```
go test ./... 
```

And check for race conditions with

```
go test ./... -race 
```

## Execution

Run the project with

```
go run . run EXECUTABLE ARGS
```

For example

```
go run . run ./test.sh 5 1
```
