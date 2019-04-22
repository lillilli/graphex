SERVICE_NAME 			:= websocket_api
PKG            		:= github.com/lillilli/graphex
PKG_LIST       		:= $(shell go list ${PKG}/... | grep -v /vendor/)
CONFIG         		:= $(wildcard local.yml)
NAMESPACE	   			:= "default"

all: setup test build

setup: ## Installing all service dependencies.
	@echo "Setup..."
	GO111MODULE=on go mod vendor

.PHONY: config
config: ## Creating the local config yml.
	@echo "Creating local config yml ..."
	cp config.example.yml local.yml

build: ## Build the executable http api file of service.
	@echo "Building..."
	cd cmd/$(SERVICE_NAME) && go build

build_scratch: ## Build all executable files for scratch.
	echo "Building for scratch..."
	cd cmd/$(SERVICE_NAME) && CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o ${SERVICE_NAME}

run: build ## Run service http api with local config.
	@echo "Running..."
	cd cmd/$(SERVICE_NAME) && ./$(SERVICE_NAME) -config=../../local.yml

clean: ## Cleans the temp files and etc.
	@echo "Clean..."
	rm -f cmd/$(SERVICE_NAME)/$(SERVICE_NAME)

lint: ## Run lint for all packages.
	echo "Linting..."
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

	golangci-lint run --enable-all --disable gochecknoglobals \
	--disable lll -e "weak cryptographic primitive" -e InsecureSkipVerify \
	--disable dupl --print-issued-lines=false

run\:image: # Run image.
	echo "Running ..."
	docker build -t graphex .
	docker stop graphex_instance || true && docker rm -f graphex_instance || true
	docker run -p 8081:8081 --name graphex_instance -v  $(PWD)/shared:/root/shared graphex

deploy: # Deploy docker image to docker hub.
	echo "Deploying ..."
	docker build -t graphex .
	docker tag graphex lillilli/graphex:latest
	docker push lillilli/graphex:latest

help: ## Display this help screen
	grep -E '^[a-zA-Z_\-\:]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ": .*?## "}; {gsub(/[\\]*/,""); printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'