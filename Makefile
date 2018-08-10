.PHONY: build test clean deps lint gofmt govet godiff coverage

BINARY = bin/forge

build: $(BINARY)

$(BINARY): deps
	CGO_ENABLED=0 go build -o '$(BINARY)'

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

BREW_FORMULA := homebrew/forge.rb
brew-release:
ifndef FORGE_VERSION
	$(eval FORGE_VERSION  := $(shell '$(BINARY)' --version | awk '{ print $$NF }'))
endif
	$(eval ARCHIVE_URL    := https://github.com/nathandines/forge/archive/$(FORGE_VERSION).tar.gz)
	$(eval ARCHIVE_SHA256 := $(shell curl -o - -Ls '$(ARCHIVE_URL)' | shasum -a 256 | awk '{ print $$1 }'))
	[ -d homebrew ] || git clone 'git@github.com:nathandines/homebrew-tap.git' 'homebrew'
	sed 's;{{ archive_url }};$(ARCHIVE_URL);g;s;{{ archive_sha256 }};$(ARCHIVE_SHA256);g' \
		homebrew_formula.rb.template > '$(BREW_FORMULA)'
	brew audit --strict '$(BREW_FORMULA)'
	cd 'homebrew' && \
		git commit -m 'forge: $(FORGE_VERSION)' -- forge.rb && \
		git push origin master
