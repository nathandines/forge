.PHONY: test clean deps lint gofmt govet godiff coverage

bin/forge:
	CGO_ENABLED=0 go build -o bin/forge

test:
	@echo "Testing forgelib"
	@cd forgelib && \
		go test -v -cover -race
	@echo "Testing commands"
	@cd commands && \
		go test -v -race

coverage:
	@cd forgelib && \
		go test -coverprofile=../coverage.out
	go tool cover -html=coverage.out

deps:
	dep ensure -v -vendor-only

clean:
	rm -rf bin

lint:
	$(MAKE) -k gofmt govet godiff

GOFMT_CMD = gofmt -s -d main.go forgelib commands
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
