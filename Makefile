DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
LDFLAGS=-ldflags="-X '$(versioningPath).buildTime=$(shell date)' -X '$(versioningPath).lastCommit=$(shell git rev-parse HEAD)' -X '$(versioningPath).semanticVersion=$(shell git describe --tags --dirty=-dev 2>/dev/null || git rev-parse --abbrev-ref HEAD)' -X '$(versioningPath).nodeVersion=$(shell go list -m all | grep celestia-node | cut -d" " -f2)'"

# Define all_pkgs, unit_pkgs, run, and cover vairables for test so that we can override them in
# the terminal more easily.
all_pkgs := $(shell go list ./...)
unit_pkgs := ./celestia
run := .
count := 1

build:
	@echo "--> Building celestia-da"
	@go build -o build/ ${LDFLAGS} ./cmd/celestia-da
.PHONY: build

## help: Show this help message
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

## clean: clean testcache
clean:
	@echo "--> Clearing testcache"
	@go clean --testcache
	rm -rf build
.PHONY: clean

## cover: generate to code coverage report.
cover:
	@echo "--> Generating Code Coverage"
	@go install github.com/ory/go-acc@latest
	@go-acc -o coverage.txt $(unit_pkgs)
.PHONY: cover

## deps: Install dependencies
deps:
	@echo "--> Installing dependencies"
	@go mod download
#	@make proto-gen
	@go mod tidy
.PHONY: deps

## lint: Run linters golangci-lint and markdownlint.
lint: vet
	@echo "--> Running golangci-lint"
	@golangci-lint run
	@echo "--> Running markdownlint"
	@markdownlint --config .markdownlint.yaml '**/*.md'
	@echo "--> Running hadolint"
	@hadolint docker/mockserv.Dockerfile
	@echo "--> Running yamllint"
	@yamllint --no-warnings . -c .yamllint.yml

.PHONY: lint

## fmt: Run fixes for linters. Currently only markdownlint.
fmt:
	@echo "--> Formatting markdownlint"
	@markdownlint --config .markdownlint.yaml '**/*.md' -f
.PHONY: fmt

## vet: Run go vet
vet:
	@echo "--> Running go vet"
	@go vet $(all_pkgs)
.PHONY: vet

## test: Running all tests
test: vet
	@echo "--> Running all tests"
	@go test -v -race -covermode=atomic -coverprofile=coverage.txt $(all_pkgs) -run $(run) -count=$(count)
.PHONY: test

## test: Running unit tests
test-unit: vet
	@echo "--> Running unit tests"
	@go test -v -race -covermode=atomic -coverprofile=coverage.txt $(unit_pkgs) -run $(run) -count=$(count)
.PHONY: test-unit

### proto-gen: Generate protobuf files. Requires docker.
#proto-gen:
#	@echo "--> Generating Protobuf files"
#	./proto/get_deps.sh
#	./proto/gen.sh
#.PHONY: proto-gen
#
### proto-lint: Lint protobuf files. Requires docker.
#proto-lint:
#	@echo "--> Linting Protobuf files"
#	@$(DOCKER_BUF) lint --error-format=json
#.PHONY: proto-lint
