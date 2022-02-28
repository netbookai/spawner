#!/usr/bin/env sh
PROTO_DIR="./proto/netbookdevs/spawnerservice"

protoc $PROTO_DIR/spawner.proto --go_out=. --go_opt=paths=source_relative
protoc $PROTO_DIR/spawner.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative
