#!/usr/bin/env sh

# Install proto3 from source
#  git clone https://github.com/google/protobuf
#  ./autogen.sh ; ./configure ; make ; make install
#
# Install proto
# Update protoc Go bindings via
#  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#
# See also
#  https://github.com/grpc/grpc-go/tree/master/examples

PROTO_DIR="./proto/netbookdevs/spawnerservice"

protoc $PROTO_DIR/spawner.proto --go_out=. --go_opt=paths=source_relative
protoc $PROTO_DIR/spawner.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative
