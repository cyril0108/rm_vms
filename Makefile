# ==========================================
# NVR Project Makefile
# ==========================================

# Docker Build Name
DOCKER_IMAGE_NAME = vmsnvr
VERSION = 0.1.52
GIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_TIME ?= $(shell date '+%Y-%m-%dT%H:%M:%S' 2>/dev/null)


# -- Output Binaries --
# The C++ worker binary name (must match what main.go looks for)
WORKER_BIN_NAME := nvr_worker
# The Go manager binary name
SERVICE_BIN_NAME := nvr_service

# -- Directories --
GO_DIR := nvr_core
CPP_ENGINE_DIR := cpp_engine
CPP_ENGINE_BUILD_DIR := $(CPP_ENGINE_DIR)/build
VUE_DIR := web

# -- Build Flags --
# -s disables symbol table, -w disables DWARF generation. Shrinks binary by ~25%
GO_LDFLAGS := -ldflags="-s -w -X 'nvr_core/buildinfo.Version=$(VERSION)' -X 'nvr_core/buildinfo.CommitSHA=$(GIT_HASH)' -X 'nvr_core/buildinfo.BuildTime=$(BUILD_TIME)'"

# -- Tools --
GO := go
CMAKE := cmake
MKDIR := mkdir -p
RM := rm -rf
CP := cp

# -- Targets --
.PHONY: all help clean build-cpp build-go docker-build docker-run export gitcommitbin docker dockersave

.DEFAULT_GOAL := help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf " \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
# 	@awk -F ':.*##' 'NF==2 {printf "%-20s %s\\n", $$1, $$2}' $(MAKEFILE_LIST)

# Default target: Build everything
all: build-go build-cpp

# Build C++ Worker
# Steps: Create build dir -> Run CMake -> Run Make -> Copy binary to root
build-cpp: ## Build C++ Worker
	@echo "--- Building C++ Worker ---"
	$(MKDIR) $(CPP_ENGINE_BUILD_DIR)
	cd $(CPP_ENGINE_BUILD_DIR) && $(CMAKE) -DTHE_VERSION=${VERSION} -DGIT_HASH=$(GIT_HASH) .. && $(MAKE)
	# Copy the compiled binary from build folder to project root
	$(CP) $(CPP_ENGINE_BUILD_DIR)/nvr_worker ./$(WORKER_BIN_NAME)
	@echo "✔ C++ Worker built successfully: ./$(WORKER_BIN_NAME)"

# Build Go NVR Service
build-go: ## Build Go NVR Service
	@echo "--- Building Go NVR Service ---"
	# Ensure go.mod exists (create if missing)
	@[ -f go.mod ] || $(GO) mod init nvr-core
	$(GO) mod tidy
	cd $(GO_DIR) && $(GO) build $(GO_LDFLAGS) -o $(SERVICE_BIN_NAME)
	@echo "✔ Go Manager built successfully: ./$(SERVICE_BIN_NAME)"

swag: ## Build Vue
	cd $(GO_DIR) && swag init

# Clean Build Artifacts
clean: ## Clean Build Artifacts
	@echo "--- Cleaning ---"
	$(RM) $(CPP_ENGINE_BUILD_DIR)
	$(RM) $(WORKER_BIN_NAME)
	$(RM) $(SERVICE_BIN_NAME)
	@echo "✔ Clean complete"

vue: ## Build Vue
	cd $(VUE_DIR) && npm run build

gitcommitbin: ## Git commit the bin
	git commit --message "Commit bin $(VERSION)"

# Build Docker image
docker: ## Build Docker image
	docker build --build-arg GIT_HASH=$(GIT_HASH) --build-arg BUILD_TIME=$(BUILD_TIME) --platform linux/amd64 -t $(DOCKER_IMAGE_NAME) .
# 	docker build -t $(DOCKER_IMAGE_NAME) .

dockersave: ## export docker image
	docker save $(DOCKER_IMAGE_NAME):latest | gzip > ../nvr_image.$(VERSION).tar.gz

export: ## extract build programs to /dist
	docker create --platform linux/amd64 --name nvr-extractor $(DOCKER_IMAGE_NAME)
	docker cp nvr-extractor:/app/nvr_service ./dist/nvr_service
	docker cp nvr-extractor:/app/nvr_worker ./dist/nvr_worker
	rsync -av --delete ./web/dist/* ./dist/web
	docker rm nvr-extractor

# Run the Docker container
docker-run: ## Run the Docker container
	docker run -it --rm $(DOCKER_IMAGE_NAME)
