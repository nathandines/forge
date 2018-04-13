.PHONY: test clean deps lint gofmt govet godiff coverage

bin/stack:
	go build -o bin/forge

test:
	@cd forgelib && \
		go test -v -cover -race

coverage:
	@cd forgelib && \
		go test -coverprofile=../coverage.out
	go tool cover -html=coverage.out

deps:
	go get -v -d -t ./...

clean:
	rm -rf bin

lint:
	$(MAKE) -k gofmt govet godiff

GOFMT_CMD = gofmt -s -d .
gofmt:
	@echo '$(GOFMT_CMD)'
	@fmtout="$$($(GOFMT_CMD))"; \
	if [ "$${fmtout:+x}" = "x" ]; then \
		echo "$$fmtout"; \
		exit 1; \
	fi

govet:
	go vet ./...

godiff:
	go tool fix -diff .
