fmt# Job Scheduler Library

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