.DEFAULT_GOAL := help

.PHONY: vm

ifeq ($(VM_DEBUG),true)
    GO_TAGS = -tags vm_debug
    VM_TARGET = debug
else
    GO_TAGS =
    VM_TARGET = all
endif

ifeq ($(shell uname -s),Darwin)
	export CGO_LDFLAGS=-framework Foundation -framework SystemConfiguration
endif




OAPI_CODEGEN=go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen


API_REST_SPEC=./openapi/openapi.yaml
API_REST_CODE_GEN_LOCATION=./openapi/generated/oapigen.go
API_REST_DOCO_GEN_LOCATION=./openapi/generated/doc.html

all: generated
generated: oapi-doc oapi-go

npm-install:
	npm init -y
	npm install

# Open API Makefile targets
oapi-validate:
	./node_modules/.bin/oas-validate -v ${API_REST_SPEC}

oapi-go: oapi-validate
	${OAPI_CODEGEN} --package oapigen --generate types,spec -o ${API_REST_CODE_GEN_LOCATION} ${API_REST_SPEC}

oapi-doc: oapi-validate
	./node_modules/.bin/redoc-cli bundle ${API_REST_SPEC} -o ${API_REST_DOCO_GEN_LOCATION}

crw: ## compile
	go build $(GO_TAGS)  -o crw/main ./cmd

clean-testcache:
	go clean -testcache

test: clean-testcache   ## tests
	go test $(GO_TAGS) ./...

install-gofumpt:
	go install mvdan.cc/gofumpt@latest

install-golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2


test-cover: vm  ## tests with coverage
	mkdir -p coverage
	go test $(GO_TAGS) -coverpkg=./... -coverprofile=coverage/coverage.out -covermode=atomic ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html


lint:
	@which golangci-lint || make install-golangci-lint
	golangci-lint run -out-format=github-actions --timeout=10m

tidy: ## add missing and remove unused modules
	 go mod tidy

format: ## run go formatter
	gofumpt -l -w .

clean: ## clean project builds
	@rm -rf ./build

help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'