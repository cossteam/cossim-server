PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
DIR := $(shell pwd)

# 将命令行参数存储到一个变量中
CMD_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
ifeq ($(strip $(CMD_ARGS)),)
    CMD_ARGS := ""
endif
SERVICE := $(lastword $(CMD_ARGS))
DOCKER_BUILD_PATH := "cmd/${SERVICE}/main.go"
INTERFACE_LIST ?=group msg relation storage user live

GOPROXY=https://goproxy.cn
TAG ?= latest
IMG ?= hub.hitosea.com/cossim/${SERVICE}:${TAG}
BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_COMMIT := ${shell git rev-parse HEAD}
BUILD_TIME := ${shell date '+%Y-%m-%d %H:%M:%S'}
BUILD_GO_VERSION := $(shell go version | grep -o  'go[0-9].[0-9].*')
VERSION_PATH := "github.com/cossim/coss-server/pkg/version"

# 防止命令行参数被误认为是目标
%:
	@:

.PHONY: dep
dep: ## Get the dependencies
	@go mod tidy

.PHONY: lint
lint: ## Lint Golang files
	@golint -set_exit_status ${PKG_LIST}

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: test
test: fmt vet## Run unittests
	@go test -short ./...


SWAG_CMD := swag i -g http.go -dir internal/push/interface/http,internal/push/api/http/model,internal/admin/interface/http,internal/admin/api/http/model,internal/group/interfaces,internal/group/api/http/model,internal/user/interface/http,internal/user/api/http/model,internal/relation/interface/http,internal/relation/api/http/model,internal/msg/interface/http,internal/msg/api/http/model,internal/storage/interface/http,internal/storage/api/http/model,internal/live/api/dto,internal/live/interface/http,pkg/utils/usersorter --packageName coss

.PHONY: swag
swag: ## Generate Swagger documentation
	$(SWAG_CMD)

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
docker-build: dep test## Build docker image with the manager.
	docker build --platform $(PLATFORMS)  --build-arg BUILD_BRANCH="${BUILD_BRANCH}" \
             --build-arg BUILD_COMMIT="${BUILD_COMMIT}" \
             --build-arg BUILD_TIME="${BUILD_TIME}" \
             --build-arg BUILD_GO_VERSION="${BUILD_GO_VERSION}" \
             --build-arg BUILD_PATH=${DOCKER_BUILD_PATH} \
             --build-arg VERSION_PATH=${VERSION_PATH} \
             --build-arg GOPROXY="${GOPROXY}" \
             -t "${IMG}" .

docker-push:
	docker push ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
#PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
PLATFORMS ?= linux/amd64
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	#sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --load --platform $(PLATFORMS) --build-arg BUILD_BRANCH="${BUILD_BRANCH}" \
             --build-arg BUILD_COMMIT="${BUILD_COMMIT}" \
             --build-arg BUILD_TIME="${BUILD_TIME}" \
             --build-arg BUILD_GO_VERSION="${BUILD_GO_VERSION}" \
             --build-arg VERSION_PATH=${VERSION_PATH} \
             --build-arg BUILD_PATH=${DOCKER_BUILD_PATH} \
             -t "${IMG}" -f Dockerfile .
	- docker buildx rm project-v3-builder
	#rm Dockerfile.cross