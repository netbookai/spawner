#!/usr/bin/env sh

# Install proto3 from source
#  git clone https://github.com/google/protobuf
#  ./autogen.sh ; ./configure ; make ; make install
#
# Update protoc Go bindings via
#  go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
#
# See also
#  https://github.com/grpc/grpc-go/tree/master/examples

# protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative spawnersvc.proto
protoc spawnersvc.proto --go_out=. --go_opt=paths=source_relative
protoc spawnersvc.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative