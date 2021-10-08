# spawner-service

# Prerequisites
1. Go needs to be installed on the system version 1.17 or above (tested on 1.17)
2. protoc needs to be installed
    ```
    apt install -y protobuf-compiler
    ```
3. protoc-gen-go and protoc-gen-go-grpc plugins to protoc needs to be installed to generate Go and gRPC code
    ```
    go get -u google.golang.org/protobuf/cmd/protoc-gen-go
    go install google.golang.org/protobuf/cmd/protoc-gen-go
    
    go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
    ```

# Running the service
1. cd into pb folder and run the compile.sh file to generate Go and gRPC generated code
2. cd into cmd/spawnersvc and run the spawnersvc.go to start the server
    ```
    go run spawnersvc.go -grpc-addr=:8083 -debug-addr=:8081
    ```
    This starts a gRPC server running on port 8083 and binds the service to it
3. cd cmd/spawnercli and run the spawnercli.go to use the client to call the service
    ```
    go run spawncli.go -grpc-addr=:8083 -method=ClusterStatus
    ```
    This calls the ClusterStatus method on the gRPC service on port 8083

# Creating a docker image

1. Build docker image from projet root directory
    ```
    docker build -t spawnerservice:0.0.1 .
    ```