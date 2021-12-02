.ONESHELL:

.PHONY: test build run docker-build docker-run

test:
	go test ${VERBOSE} ./... -race

build:
	go build .

run:
	go run ${VERBOSE} . ${ARGS}

docker-build:
	docker build . -t golang-dev

docker-run:
	docker run -it golang-dev
