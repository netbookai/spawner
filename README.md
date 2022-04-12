# spawner-service

# Prerequisites
1. Go needs to be installed on the system version 1.17 or above (tested on 1.17)
2. protoc needs to be installed
    ```
    apt install -y protobuf-compiler
    ```
3. protoc-gen-go and protoc-gen-go-grpc plugins to protoc needs to be installed to generate Go and gRPC code
    ```
    go install google.golang.org/protobuf/cmd/protoc-gen-go
    
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
    ```



# Development

1. Run make proto to generate rpc and protobuf go files
    ```
    make proto
    ```
2. setup AWS_ROLE_ARN environment variable
    ```
    export AWS_ROLE_ARN=arn:aws:iam::965734315247:role/sandboxClusterSecretManagerRole
    ```
3. cd into cmd/spawnersvc and run the spawnersvc.go to start the server
    ```
    make run
    ```
    This starts a gRPC server running on port 8083 and binds the service to it
4. cd cmd/spawnercli and run the spawnercli.go to use the client to call the service
    ```
    go run spawncli.go -grpc-addr=:8083 -method=ClusterStatus
    ```
    This calls the ClusterStatus method on the gRPC service on port 8083
5. To update modules,
    ```
    make tidy
    ```
6. Run formatter before committing or set your editor to FORMAT ON SAVE with goimports.
    ```
    make fmt
    ```

# Creating a docker image

1. Build docker image from projet root directory
    ```
    docker login registry.gitlab.com
    docker build -t registry.gitlab.com/netbook-devs/spawner-service/spawnerservice:0.0.1 .
    docker push registry.gitlab.com/netbook-devs/spawner-service/spawnerservice:0.0.1
    ```

# Running the app using helm

1. (Optional) Create  a new docker registry secret from docker config file
    ```
    # base64 encode username and password
    echo -n <username>:<password> | base64
    # base64 encode ~/.docker/config.json file
    cat ~/.docker/config.json | base64
    ```
1. Install the helm chart
    ```
    helm install spawnerservice kubernetes/charts/spawnerservice -f kubernetes/charts/spawnerservice/deployments/dev/spawnerservice.yaml
    ```
    Service will be running at `spawnerservice-service:80` inside k8s cluster

2. Test server deployment
    ```
    kubectl exec -it spawner-cli -- /bin/sh
    ./spawnercli -grpc-addr=spawnerservice-service:80 -method=ClusterStatus
    ```



# spawner-cli

## build and install

### build

```
make build-client
```

### install

The above command will generate the client binary named `spawner-cli` in the current working directory. Copy that to your path or use it with relative execution path `./spawner-cli` as per your convenience.

## Usage

For all the commands you need to pass spawner host address, default value is set to `localhost:8083`, if you need to change that, pass in using `--addr` or `-a`.

Example:

```
spawner-cli cluster-status clustername --addr=192.168.1.78:8080 --provider=aws --region=us-west-2
```

### Create a new cluster

To create a cluster we need more information on the cluster and node specification which can be passed to command as a file by specifying `--request` or `-r`

```
spawner-cli create-cluster clustername -r request.json
```

request.json should contain the following

```
{
  "provider": "aws",
  "region": "us-east-1",
  "labels": {
    "fugiat8f": "laboris magna Duis amet"
  },
  "node": {
    "name": "proident",
    "diskSize": 10,
    "labels": {
      "laboris48": "velit aute in eiusmod",
      "in_0": "in incididunt do nostrud",
    },
    "instance": "standard_A2_V2",
    "gpuEnabled": false
  }
}

```

> Note : This wil create a cluster and attach new node to it as per spec, the time taken by this operation completely depends on how fast provider responds.

---

### Cluster status

Get the cluster status such as CREATING, ACTIVE, DELETING

```
spawner-cli cluster-status clustername --provider "aws" -r=region
```
----

### Delete Cluster 

Delete the existing cluster
```
spawner-cli delete-cluster clustername --provider "aws" -r=region
```

If the cluster has the nodes attached to it, this operation will fail, you can force delete the cluster which deletes attached node and then deletes the cluster.

To force delete set the `--force` or `-f`

```
spawner-cli delete-cluster clustername --provider "aws" -r=region --force
```
