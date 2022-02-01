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
