ALL_GO_FILES=$(shell find . -type f  -name '*.go')
GO_VERSION=1.17


deps:
	go mod tidy -compat=$(GO_VERSION)
run:
	go run cmd/spawnersvc/main.go
test:
	go test ./...

clean:
	go clean ./...

.PHONY: proto
proto:
	./proto/compile.sh

fmt:
	goimports -w $(ALL_GO_FILES)

lint:
	golint ./...
