PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

MAIN_FILE=cmd/main.go
NAME= ""
DIR := $(shell pwd)
IMG ?= hub.hitosea.com/coss-server/${ACTION}-${NAME}:latest

BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
#BUILD_BRANCH := "main"
BUILD_COMMIT := ${shell git rev-parse HEAD}
#BUILD_COMMIT := "22193944514397212fe6a25189906ab9de49164a"
BUILD_TIME := ${shell date '+%Y-%m-%d %H:%M:%S'}
BUILD_GO_VERSION := $(shell go version | grep -o  'go[0-9].[0-9].*')
VERSION_PATH := "github.com/cossim/coss-server/pkg/version"
BUILD_PATH := ""
DOCKER_BUILD_PATH := ""
ACTION := ""

# 根据传入的 ACTION 参数设置 BUILD_PATH
ifeq ($(ACTION), interfaces)
	BUILD_PATH := ${DIR}/interfaces/${NAME}
	DOCKER_BUILD_PATH :="interfaces/${NAME}"
else ifeq ($(ACTION), services)
	BUILD_PATH := ${DIR}/services/${NAME}
	DOCKER_BUILD_PATH := "services/${NAME}"
endif

# 如果没有设置 BUILD_PATH，输出错误信息
ifeq ($(BUILD_PATH),)
    $(error Invalid ACTION. Use 'make build ACTION=interfaces' or 'make build ACTION=services')
endif

.PHONY: dep test build-service build-interface docker-build docker-push

dep: ## Get the dependencies
	@go mod tidy

test: ## Run unittests
	@go test -short ${PKG_LIST}

# 构建指定grpc服务  make build-services ACTION=services NAME="user"
build-services: dep ## Build the binary file
ifdef NAME
	@echo "Building with flags: go build -ldflags \"-s -w\" -ldflags \"-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' -X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' -X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' -X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}'\" -o ${BUILD_PATH}/$(MAIN_FILE)"
	@go build -ldflags "-s -w" -ldflags "-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' -X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' -X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' -X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}'" -o ${BUILD_PATH}/bin/main ${BUILD_PATH}/$(MAIN_FILE)
else
	@echo "Please provide service NAME"
endif

# 构建指定接口服务  make build-interfaces ACTION=interfaces NAME="user"
build-interfaces: dep
ifdef NAME
	@echo "Building ${INTERFACE_NAME} interface with flags: go build -ldflags \"-s -w\" -ldflags \"-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' -X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' -X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' -X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}'\" -o ${BUILD_PATH}/$(MAIN_FILE)"
	@go build -ldflags "-s -w" -ldflags "-X '${VERSION_PATH}.GitBranch=${BUILD_BRANCH}' -X '${VERSION_PATH}.GitCommit=${BUILD_COMMIT}' -X '${VERSION_PATH}.BuildTime=${BUILD_TIME}' -X '${VERSION_PATH}.GoVersion=${BUILD_GO_VERSION}'" -o ${BUILD_PATH}/bin/main ${BUILD_PATH}/$(MAIN_FILE)
else
	@echo "Please provide interface NAME"
endif

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
docker-build: ## Build docker image with the manager.
	#docker build -t ${IMG} .
	# 根据传入的 ACTION 参数设置 BUILD_PATH
	docker build --build-arg BUILD_BRANCH="${BUILD_BRANCH}" \
             --build-arg BUILD_COMMIT="${BUILD_COMMIT}" \
             --build-arg BUILD_TIME="${BUILD_TIME}" \
             --build-arg BUILD_GO_VERSION="${BUILD_GO_VERSION}" \
             --build-arg BUILD_PATH="${DOCKER_BUILD_PATH}" \
             --build-arg VERSION_PATH="${VERSION_PATH}" \
              --build-arg MAIN_FILE="${MAIN_FILE}" \
             -t "${IMG}" .

docker-push: ## Push docker image with the manager.
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