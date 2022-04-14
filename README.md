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


# Releases
 ### binary release - coming soon
 ### source release -clone repo
 ```
 git clone git@gitlab.com:netbook-devs/spawner-service.git
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
3. run the server using
    ```
    make run
    ```
    This starts a gRPC server running on port 8083 and binds the service to it

4. To update modules,
    ```
    make tidy
    ```
5. Run formatter before committing or set your editor to FORMAT ON SAVE with goimports.
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



# spawner command line tool

## build and install

### build

```
make build-client
```

### install

The above command will generate the client binary named `spawner` in the current working directory. Copy that to your path or use it with relative execution path `./spawner` as per your convenience.

## Usage

For all the commands you need to pass spawner host address, default value is set to `localhost:8083`, if you need to change that, pass in using `--addr` or `-a`.

Example:

```
spawner cluster-status clustername --addr=192.168.1.78:8080 --provider=aws --region=us-west-2
```

### Create a new cluster

To create a cluster we need more information on the cluster and node specification which can be passed to command as a file by specifying `--request` or `-r`

```
spawner create-cluster clustername -r request.json
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
spawner cluster-status clustername --provider "aws" -r=region
```
----

### Delete Cluster 

Delete the existing cluster
```
spawner delete-cluster clustername --provider "aws" -r=region
```

If the cluster has the nodes attached to it, this operation will fail, you can force delete the cluster which deletes attached node and then deletes the cluster.

To force delete set the `--force` or `-f`

```
spawner delete-cluster clustername --provider "aws" -r=region --force
```

### Add new nodepool
Create new nodepool in a given cluster

```
spawner nodepool add clustername --request request.json
```

request.json will contain the nodespec for the new nodepool,

```
@request.json

{
  "nodeSpec": {
    "diskSize": 31,
    "name": "prosint",
    "count": 3,
    "instance": "Standard_A2_V2",
    "labels": {
      "created_by": "cli"
    }
  },
  "region": "eastus2",
  "clusterName": "my-cluster",
  "provider": "azure"
}
```
---

### Delete nodepool

```
spawner nodepool delete clustername --provider "aws" -r=region --nodepool nodepoolname
```

---

### Get kubeconfg for the cluster
```
spawner kubeconfig clustername --provider "aws" -r=region
```

this will read existing kube config from `~/.kube/config` and merges new cluster config to it, sets the current context as the requested cluster
