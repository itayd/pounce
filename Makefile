.PHONY: all
all: pounce test

pounce: go.* *.go
	go build -ldflags="-X 'main.version=dev-$(shell git rev-parse HEAD)'"

.PHONY: test
test:
	./tests/run.sh

.PHONY: install
install: pounce
	go install
