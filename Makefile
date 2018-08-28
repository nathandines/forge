.PHONY: build test clean update-deps lint gofmt govet godiff coverage choco-package choco-release brew-release

BINARY = bin/forge

build: $(BINARY)

$(BINARY):
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

update-deps:
	go get -u
	go mod tidy

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
	$(eval FORGE_VERSION  ?= $(shell '$(BINARY)' --version | awk '{ print $$NF }'))
	$(eval ARCHIVE_URL    := https://github.com/nathandines/forge/archive/$(FORGE_VERSION).tar.gz)
	$(eval ARCHIVE_SHA256 := $(shell curl -o - -Ls '$(ARCHIVE_URL)' | shasum -a 256 | awk '{ print $$1 }'))
	[ -d homebrew ] || git clone 'git@github.com:nathandines/homebrew-tap.git' 'homebrew'
	sed -e 's;{{ archive_url }};$(ARCHIVE_URL);g' \
		-e 's;{{ archive_sha256 }};$(ARCHIVE_SHA256);g' \
		homebrew_formula.rb.template > '$(BREW_FORMULA)'
	brew audit --strict '$(BREW_FORMULA)'
	cd 'homebrew' && \
		git commit -m 'forge: $(FORGE_VERSION)' -- forge.rb && \
		git push origin master

CHOCO            ?= choco
CHOCO_FORGE_PATH := $(PWD)/chocolatey/forge
CHOCO_NUSPEC     := $(CHOCO_FORGE_PATH)/forge.nuspec
LICENSE_COMMIT   ?= master
choco-package:
	$(eval FORGE_VERSION ?= $(shell '$(BINARY)' --version | awk '{ print $$NF }'))
	$(eval URL_x64 := https://github.com/nathandines/forge/releases/download/$(FORGE_VERSION)/forge_$(FORGE_VERSION)_windows_amd64.exe)
	$(eval URL_386 := https://github.com/nathandines/forge/releases/download/$(FORGE_VERSION)/forge_$(FORGE_VERSION)_windows_386.exe)
	mkdir -p '$(CHOCO_FORGE_PATH)/tools'
	rm -f '$(CHOCO_FORGE_PATH)/tools/forge_'*.zip
	cp LICENSE '$(CHOCO_FORGE_PATH)/tools/LICENSE.txt'
	sed 's/{{ package_version }}/$(FORGE_VERSION:v%=%)/g' 'chocolatey/forge/chocolatey_package.nuspec.template' > '$(CHOCO_NUSPEC)'
	curl -Lo 'forge64.exe' '$(URL_x64)'
	curl -Lo 'forge32.exe' '$(URL_386)'
	zip -v '$(CHOCO_FORGE_PATH)/tools/forge_$(FORGE_VERSION).zip' forge64.exe forge32.exe
	sed -e 's;{{ url_x64 }};$(URL_x64);g' \
		-e "s;{{ sha256_x64 }};$$(shasum -a 256 forge64.exe | awk '{ print $$1 }');g" \
		-e 's;{{ url_386 }};$(URL_386);g' \
		-e "s;{{ sha256_386 }};$$(shasum -a 256 forge32.exe | awk '{ print $$1 }');g" \
		-e 's;{{ license_commit }};$(LICENSE_COMMIT);g' \
		'chocolatey/forge/VERIFICATION.txt.template' > 'chocolatey/forge/tools/VERIFICATION.txt'
	rm -f forge32.exe forge64.exe '$(CHOCO_FORGE_PATH)/forge.$(FORGE_VERSION:v%=%).nupkg'
	cd '$(CHOCO_FORGE_PATH)' && $(CHOCO) pack '$(CHOCO_NUSPEC)'

choco-release: choco-package
	$(eval FORGE_VERSION ?= $(shell '$(BINARY)' --version | awk '{ print $$NF }'))
	@echo 'Pushing to Chocolatey...'
	@dotnet nuget push -k '$(CHOCO_API_KEY)' -s 'https://push.chocolatey.org/' '$(CHOCO_FORGE_PATH)/forge.$(FORGE_VERSION:v%=%).nupkg'
