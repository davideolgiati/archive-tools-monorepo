# Change these variables as necessary.
main_package_path = ./cmd/example
binary_name = example

export TMPDIR = $(shell pwd)/bin/golang_builds/

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | awk -F ':' '{printf "%-20s %s\n", $1, $2}' | sed -e 's/^/ /'

setup-env:
	mkdir -p ${TMPDIR}

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell go run mvdan.cc/gofumpt@latest -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
##go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
.PHONY: test
test: setup-env
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover: setup-env
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -func /tmp/coverage.out

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## build: build the application
.PHONY: build
build: setup-env
	go build -o=/tmp/bin/${binary_name} ${main_package_path}
