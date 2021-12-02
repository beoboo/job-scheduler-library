# Job Scheduler Library

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

## Daemon

The project contains an [example](cmd/daemon/main.go) of an executable that can answer to "child" commands.

Build it with:

```
go build -o bin/ ./cmd/daemon
```

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
