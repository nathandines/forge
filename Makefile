.PHONY: test clean deps lint gofmt govet godiff

bin/stack:
	go build -o bin/stack

test:
	@cd stacklib && \
		go test -v -cover -race

deps:
	go get -v -d -t ./...

clean:
	rm -rf bin

lint:
	$(MAKE) -k gofmt govet godiff

gofmt:
	@echo 'gofmt -d .'
	@fmtout="$$(gofmt -d .)"; \
	if [ "$${fmtout:+x}" = "x" ]; then \
		echo "$$fmtout"; \
		exit 1; \
	fi

govet:
	go vet ./...

godiff:
	go tool fix -diff .
