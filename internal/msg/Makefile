PROJECT_NAME=msg
MAIN_FILE=cmd/main.go
PKG := "github.com/cossim/coss-server/internal/$(PROJECT_NAME)"
MOD_DIR := $(shell go env GOMODCACHE)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep lint vet test test-coverage build clean

BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_COMMIT := ${shell git rev-parse HEAD}
BUILD_TIME := ${shell date '+%Y-%m-%d %H:%M:%S'}
BUILD_GO_VERSION := $(shell go version | grep -o  'go[0-9].[0-9].*')
VERSION_PATH := "github.com/cossim/coss-server/pkg/version"

all: build

dep: ## Get the dependencies
	@go mod tidy

lint: ## Lint Golang files
	@golint -set_exit_status ${PKG_LIST}

vet: ## Run go vet
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST}
	@cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@cd PKG
	@#echo "Building with flags: go build -ldflags \"-s -w\" -ldflags \"-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' -X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' -X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' -X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}'\" -o bin/main $(MAIN_FILE)"
	@go build -ldflags "-s -w" -ldflags "-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' -X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' -X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' -X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}'" -o bin/main $(MAIN_FILE)

build-linux: dep ## Build the binary file
	@GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -ldflags "-X '${VERSION_PATH}.GIT_BRANCH=${BUILD_BRANCH}' -X '${VERSION_PATH}.GIT_COMMIT=${BUILD_COMMIT}' -X '${VERSION_PATH}.BUILD_TIME=${BUILD_TIME}' -X '${VERSION_PATH}.GO_VERSION=${BUILD_GO_VERSION}'" -o bin/main $(MAIN_FILE)

gen: gen-grpc gen-openapi

.PHONY: gen-grpc
gen-grpc:
	-@protoc -I=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false api/grpc/v1/*.proto
	-@go fmt ./...
	-@protoc-go-inject-tag -input=api/grpc/v1/*.pb.go

.PHONY: gen-openapi
gen-openapi:
	-@oapi-codegen -generate gin,spec -o api/http/v1/openapi.gen.go -package v1 api/http/v1/*.yaml
	-@oapi-codegen -generate types -o api/http/v1/openapi.types.go -package v1 api/http/v1/*.yaml

run: # Run Develop server
	@go run ../../cmd/$(PROJECT_NAME)/main.go --config ./config/config.yaml

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
