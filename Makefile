GO_VERSION=1.17

ALL_GO_FILES=$(shell find . -type f  -name '*.go')
ALL_PROTO_FILES=$(shell find ./proto/netbookdevs -type f  -name '*.proto')
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

install:
	@helm upgrade --install spawnerservice kubernetes/charts/spawnerservice -f kubernetes/charts/spawnerservice/deployments/dev/spawnerservice.yaml --set docker=$(DOCKER_K8S_CONFIG),rancher.address=$(RANCHER_ADDRESS),rancher.username=$(RANCHER_USERNAME),rancher.password=$(RANCHER_PASSWORD),rancher.aws_cred_name=$(RANCHER_AWS_CRED_NAME),image.tag=$(TAG),env=$(ENV),secret_host_region=$(SECRET_HOST_REGION),route53_hostedzone_id=$(AWS_ROUTE53_HOSTEDZONEID_DEV),node_deletion_timeout_in_seconds=$(NODE_DELETION_TIME_IN_SECONDS),azure_cloud_provider=$(AZURE_CLOUD_PROVIDER)