SERVER_NAME="gophermart"

# .PHONY: build
.PHONY: all dep build clean test race lint

all: build

lint: ## Lint the files
	@ golangci-lint run --config .golangci.yml ./...

test: ## Run unit  tests
	# @go test -short ${PKG_LIST}
	@go test -v -race -timeout 30s ./...

race: dep ## Run data race detector
	@go test -race -short ./...

dep: ## Get the dependencies
	@go get -v -d ./...

build: ## Build the binary file
	@go build -o ${SERVER_NAME} -v ./cmd/gophermart

clean: ## Remove previous build
	@rm -f $(SERVER_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build