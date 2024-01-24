PROJECT_NAME=user
MAIN_FILE=cmd/main.go
PKG := "github.com/cossim/coss-server/interfaces/$(PROJECT_NAME)"
MOD_DIR := $(shell go env GOMODCACHE)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep lint vet test test-coverage build clean

BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
#BUILD_BRANCH := "main"
BUILD_COMMIT := ${shell git rev-parse HEAD}
#BUILD_COMMIT := "22193944514397212fe6a25189906ab9de49164a"
BUILD_TIME := ${shell date '+%Y-%m-%d %H:%M:%S'}
BUILD_GO_VERSION := $(shell go version | grep -o  'go[0-9].[0-9].*')
VERSION_PATH := "github.com/cossim/coss-server/pkg/version"


# Image URL to use all building/pushing image targets
IMG ?= hub.hitosea.com/coss-server:latest

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

install: ## Install depence go package
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/favadi/protoc-go-inject-tag@latest

gen: ## generate protobuf file
	@protoc -I=. --go_out=. --go_opt=module=${PKG} --go-grpc_out=. --go-grpc_opt=module=${PKG} api/v1/*.proto
	@go fmt ./...
	@protoc-go-inject-tag -input=api/v1/*.pb.go

run: # Run Develop server
	@go run $(MAIN_FILE) --config ./config/config.yaml --direct

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'



# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	# make docker-build ACTION=interface NAME=user
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	# make docker-build ACTION=interface NAME=user
	docker push ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross