package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	clusterName = "test-nsp-cluster-05"
	region      = "us-west-2"
	provider    = "aws"
)

func main() {

	var logger, _ = zap.NewDevelopment()
	var sugar = logger.Sugar()
	defer sugar.Sync()

	fs := flag.NewFlagSet("spawncli", flag.ExitOnError)
	grpcAddr := fs.String("grpc-addr", ":8083", "gRPC address of addsvc")
	method := fs.String("method", "ClusterStatus", "ClusterStatus")
	fs.Usage = usageFor(fs, os.Args[0]+" [flags] <a> <b>")
	fs.Parse(os.Args[1:])

	if *grpcAddr == "" {
		sugar.Errorf("host address is empty '%s'", *grpcAddr)
		os.Exit(1)
	}
	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		sugar.Errorw("error connecting to remote", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := proto.NewSpawnerServiceClient(conn)

	if err != nil {
		sugar.Errorw("error connecting to remote", "error", err)
		os.Exit(1)
	}

	node := &proto.NodeSpec{
		Name:     "sandbox-test-nsp-ng-01",
		Instance: "t3.medium",
		DiskSize: 13,
	}
	createClusterReq := &proto.ClusterRequest{
		Provider: provider,
		Region:   region,
		Node:     node,
		Labels: map[string]string{
			"user":        "98fe250a-7d98-4604-8317-1fbadda737ea",
			"workspaceid": "18638c97-7352-426e-a79e-241956188fed",
		},
		ClusterName: clusterName,
	}

	addTokenReq := &proto.AddTokenRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
	}

	getTokenReq := &proto.GetTokenRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
	}

	addRoute53RecordReq := &proto.AddRoute53RecordRequest{
		DnsName:    "af196cc69b2644f6480ddf353a8508d2-1819137011.us-west-1.elb.amazonaws.com",
		RecordName: "*.mani.app.netbook.ai",
		Region:     region,
		Provider:   provider,
		// RegionIdentifier: "Oregon region",
	}

	clusterStatusReq := &proto.ClusterStatusRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
	}

	getClustersReq := &proto.GetClustersRequest{
		Region:   region,
		Provider: provider,
	}

	getClusterReq := &proto.GetClusterRequest{
		ClusterName: clusterName,
		Provider:    provider,
		Region:      region,
	}

	addNode := &proto.NodeSpec{
		Name:       "sandbox-node-ng-gpu-01",
		Instance:   "t2.medium",
		DiskSize:   20,
		GpuEnabled: true,
	}

	addNodeReq := &proto.NodeSpawnRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		NodeSpec:    addNode,
	}

	deleteClusterReq := &proto.ClusterDeleteRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
	}

	deleteNodeReq := &proto.NodeDeleteRequest{
		ClusterName:   clusterName,
		NodeGroupName: "ng-04",
		Region:        region,
		Provider:      provider,
	}

	createVolumeReq := &proto.CreateVolumeRequest{
		Availabilityzone: "us-west-2a",
		Volumetype:       "gp2",
		Size:             1,
		Snapshotid:       "",
		Region:           region,
		Provider:         provider,
	}

	deleteVolumeReq := &proto.DeleteVolumeRequest{
		Volumeid: "vol-05d7e98ae385b2e29",
		Region:   region,
		Provider: provider,
	}

	createSnapshotReq := &proto.CreateSnapshotRequest{
		Volumeid: "vol-07ccb258225e0e213",
		Region:   region,
		Provider: provider,
	}
	createSnapshotAndDeleteReq := &proto.CreateSnapshotAndDeleteRequest{
		Volumeid: "vol-0f220de036ebea748",
		Region:   region,
		Provider: provider,
	}

	getWorkspaceCost := &proto.GetWorkspaceCostRequest{
		WorkspaceId: "test",
		Provider:    "aws",
		AccountName: "965734315247",
		StartDate:   "2021-08-01",
		EndDate:     "2022-02-01",
		Granularity: "MONTHLY",
		CostType:    "BlendedCost",
		GroupBy:     "SERVICE",
	}

	switch *method {
	case "CreateCluster":
		v, err := client.CreateCluster(context.Background(), createClusterReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating cluster", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateCluster method", "response", v)
	case "AddToken":
		v, err := client.AddToken(context.Background(), addTokenReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error adding token", "error", err)
			os.Exit(1)
		}
		sugar.Infow("AddToken method", "reponse", v)
	case "GetToken":
		v, err := client.GetToken(context.Background(), getTokenReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error getting token", "error", err)
			os.Exit(1)
		}
		sugar.Infow("GetToken method", "response", v)
	case "AddRoute53Record":
		v, err := client.AddRoute53Record(context.Background(), addRoute53RecordReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating Alias record", "error", err)
			os.Exit(1)
		}
		sugar.Infow("AddRoute53Record method", "response", v)
	case "GetCluster":
		v, err := client.GetCluster(context.Background(), getClusterReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		sugar.Infow("GetCluster method", "response", v)
	case "GetClusters":
		v, err := client.GetClusters(context.Background(), getClustersReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		sugar.Infow("GetClusters method", "response", v)
	case "ClusterStatus":
		v, err := client.ClusterStatus(context.Background(), clusterStatusReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error fetching cluster status", "error", err)
			os.Exit(1)
		}
		sugar.Infow("ClusterStatus method", "response", v)
	case "AddNode":
		v, err := client.AddNode(context.Background(), addNodeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error adding node", "error", err)
			os.Exit(1)
		}
		sugar.Infow("AddNode method", "response", v)
	case "DeleteCluster":
		v, err := client.DeleteCluster(context.Background(), deleteClusterReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error deleting cluster", "error", err)
			os.Exit(1)
		}
		sugar.Infow("DeleteCluster method", "response", v)
	case "DeleteNode":
		v, err := client.DeleteNode(context.Background(), deleteNodeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error deleting node", "error", err)
			os.Exit(1)
		}
		sugar.Infow("DeleteNode method", "response", v)

	case "CreateVolume":
		v, err := client.CreateVolume(context.Background(), createVolumeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating volume", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateVolume method", "response", v)

	case "DeleteVolume":
		v, err := client.DeleteVolume(context.Background(), deleteVolumeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error deleting volume", "error", err)
			os.Exit(1)
		}
		sugar.Infow("DeleteVolume method", "response", v)

	case "CreateSnapshot":
		v, err := client.CreateSnapshot(context.Background(), createSnapshotReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating snapshot", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateSnapshot method", "response", v)

	case "CreateSnapshotAndDelete":
		v, err := client.CreateSnapshotAndDelete(context.Background(), createSnapshotAndDeleteReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating snapshot and deleting volume", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateSnapshotAndDelete method", "response", v)

	case "RegisterWithRancher":
		v, err := client.RegisterWithRancher(context.Background(), &proto.RancherRegistrationRequest{
			ClusterName: clusterName,
		})
		if err != nil && err.Error() != "" {
			sugar.Errorw("error registering cluster with rancher", "error", err)
			os.Exit(1)
		}
		sugar.Infow("RegisterWithRancher method", "response", v)
	case "GetWorkspaceCost":
		v, err := client.GetWorkspaceCost(context.Background(), getWorkspaceCost)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error registering cluster with rancher", "error", err)
			os.Exit(1)
		}
		sugar.Infow("GetWorkspaceCost method", "response", v)
	default:
		sugar.Infow("error: invalid method", "method", *method)
		os.Exit(1)
	}
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")

		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
