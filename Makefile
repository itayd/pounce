.PHONY: pounce
pounce:
	go build .

.PHONY: test
test:
	./tests/run.sh
