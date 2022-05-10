GO_VERSION=1.17

ALL_GO_FILES=$(shell find . -type f  -name '*.go')
ALL_PROTO_FILES=$(shell find ./proto/ -type f  -name '*.proto')
CI_COMMIT_SHORT_SHA ?= "local"
TAG ?= $(CI_COMMIT_SHORT_SHA)


tidy:
	go mod tidy -compat=$(GO_VERSION)

run:
	go run cmd/spawnersvc/main.go

test:
	go test ./...

clean:
	go clean ./...
	rm spawner

.PHONY: proto
proto:
	@echo "generating proto code"
	@./proto/compile.sh

fmt:
	goimports -w $(ALL_GO_FILES)

lint:
	golint ./...

fmt-proto:
	clang-format --style=Chromium -i $(ALL_PROTO_FILES)

build-client:
	go build -o spawner ./cmd/client/main.go 
