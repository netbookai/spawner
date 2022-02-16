ALL_GO_FILES=$(shell find . -type f  -name '*.go')


deps:
	go mod tidy
run:
	go run cmd/spawnersvc/spawnersvc.go
test:
	go test ./...

clean:
	go clean ./...

proto:
	./pb/compile.sh

fmt:
	goimports -w $(ALL_GO_FILES)

lint:
	golint ./...
