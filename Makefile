.PHONY: run test clean deps

bin/stack:
	go build -o bin/stack

test:
	@cd stacklib && \
		go test -v -cover -race

deps:
	go get -v -d -t ./...

clean:
	rm -rf bin
