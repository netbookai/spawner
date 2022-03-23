#!/usr/bin/env sh

# find all proto files in proto directory
ALL_PROTO_FILES=$(find ./proto -type f  -name '*.proto')

for proto in $ALL_PROTO_FILES; do
    protoc $proto --go_out=. --go_opt=paths=source_relative
    protoc $proto --go-grpc_out=. --go-grpc_opt=paths=source_relative
done 

