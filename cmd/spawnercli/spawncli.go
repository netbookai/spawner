package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwntransport"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {

	var logger, _ = zap.NewDevelopment()
	var sugar = logger.Sugar()
	defer sugar.Sync()

	fs := flag.NewFlagSet("spawncli", flag.ExitOnError)
	var (
		// httpAddr = fs.String("http-addr", "", "HTTP address of addsvc")
		grpcAddr = fs.String("grpc-addr", ":8083", "gRPC address of addsvc")
		method   = fs.String("method", "ClusterStatus", "ClusterStatus")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags] <a> <b>")
	fs.Parse(os.Args[1:])

	// This is a demonstration client, which supports multiple transports.
	var (
		svc spawnerservice.ClusterController
		err error
	)

	if *grpcAddr != "" {
		conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			sugar.Errorw("error connecting to remote", "error", err)
			os.Exit(1)
		}
		defer conn.Close()
		svc = spwntransport.NewGRPCClient(conn, zap.NewNop().Sugar())
	} else {
		sugar.Errorw("error connecting to remote", "error", err)
		os.Exit(1)
	}
	if err != nil {
		sugar.Errorw("error connecting to remote", "error", err)
		os.Exit(1)
	}

	node := &pb.NodeSpec{
		Name:     "sandbox-test-nsp-ng-01",
		Instance: "t3.medium",
		DiskSize: 13,
	}
	createClusterReq := &pb.ClusterRequest{
		Provider:    "aws",
		Region:      "us-west-2",
		Node:        node,
		Labels:      map[string]string{},
		ClusterName: "sandbox-test-nsp-3",
	}

	addTokenReq := &pb.AddTokenRequest{
		ClusterName: "aws-us-west-2-eks-4",
		Region:      "us-west-2",
	}

	getTokenReq := &pb.GetTokenRequest{
		ClusterName: "sandbox-test-nsp-3",
		Region:      "us-west-2",
	}

	addRoute53RecordReq := &pb.AddRoute53RecordRequest{
		DnsName:    "af196cc69b2644f6480ddf353a8508d2-1819137011.us-west-1.elb.amazonaws.com",
		RecordName: "*.mani.app.netbook.ai",
		RegionName: "us-west-1",
		// RegionIdentifier: "Oregon region",
	}

	clusterStatusReq := &pb.ClusterStatusRequest{
		ClusterName: "sandbox-test-nsp-3",
		Region:      "us-west-2",
	}

	getClustersReq := &pb.GetClustersRequest{
		Provider: "aws",
		Region:   "us-west-2",
		Scope:    "public",
	}

	getClusterReq := &pb.GetClusterRequest{
		ClusterName: "sandbox-nsp",
	}

	addNode := &pb.NodeSpec{
		Name:       "sandbox-node-ng-gpu-01",
		Instance:   "t2.medium",
		DiskSize:   20,
		GpuEnabled: true,
	}

	addNodeReq := &pb.NodeSpawnRequest{
		ClusterName: "sandbox-test-nsp-3",
		Region:      "us-west-2",
		NodeSpec:    addNode,
	}

	deleteClusterReq := &pb.ClusterDeleteRequest{
		ClusterName: "sandbox-test-nsp-1",
		Region:      "us-west-2",
	}

	deleteNodeReq := &pb.NodeDeleteRequest{
		ClusterName:   "sandbox-test-nsp-1",
		NodeGroupName: "ng-04",
		Region:        "us-west-2",
	}

	createVolumeReq := &pb.CreateVolumeRequest{
		Availabilityzone: "us-west-2a",
		Volumetype:       "gp2",
		Size:             1,
		Snapshotid:       "",
		Provider:         "aws",
		Region:           "us-west-2",
	}

	deleteVolumeReq := &pb.DeleteVolumeRequest{
		Volumeid: "vol-05d7e98ae385b2e29",
		Provider: "aws",
		Region:   "us-west-2",
	}

	createSnapshotReq := &pb.CreateSnapshotRequest{
		Volumeid: "vol-07ccb258225e0e213",
		Provider: "aws",
		Region:   "us-west-2",
	}
	createSnapshotAndDeleteReq := &pb.CreateSnapshotAndDeleteRequest{
		Volumeid: "vol-0f220de036ebea748",
		Provider: "aws",
		Region:   "us-west-2",
	}

	switch *method {
	case "CreateCluster":
		v, err := svc.CreateCluster(context.Background(), createClusterReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating cluster", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateCluster method", "response", v)
	case "AddToken":
		v, err := svc.AddToken(context.Background(), addTokenReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error adding token", "error", err)
			os.Exit(1)
		}
		sugar.Infow("AddToken method", "reponse", v)
	case "GetToken":
		v, err := svc.GetToken(context.Background(), getTokenReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error getting token", "error", err)
			os.Exit(1)
		}
		sugar.Infow("GetToken method", "response", v)
	case "AddRoute53Record":
		v, err := svc.AddRoute53Record(context.Background(), addRoute53RecordReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating Alias record", "error", err)
			os.Exit(1)
		}
		sugar.Infow("AddRoute53Record method", "response", v)
	case "GetCluster":
		v, err := svc.GetCluster(context.Background(), getClusterReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		sugar.Infow("GetCluster method", "response", v)
	case "GetClusters":
		v, err := svc.GetClusters(context.Background(), getClustersReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		sugar.Infow("GetClusters method", "response", v)
	case "ClusterStatus":
		v, err := svc.ClusterStatus(context.Background(), clusterStatusReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error fetching cluster status", "error", err)
			os.Exit(1)
		}
		sugar.Infow("ClusterStatus method", "response", v)
	case "AddNode":
		v, err := svc.AddNode(context.Background(), addNodeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error adding node", "error", err)
			os.Exit(1)
		}
		sugar.Infow("AddNode method", "response", v)
	case "DeleteCluster":
		v, err := svc.DeleteCluster(context.Background(), deleteClusterReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error deleting cluster", "error", err)
			os.Exit(1)
		}
		sugar.Infow("DeleteCluster method", "response", v)
	case "DeleteNode":
		v, err := svc.DeleteNode(context.Background(), deleteNodeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error deleting node", "error", err)
			os.Exit(1)
		}
		sugar.Infow("DeleteNode method", "response", v)

	case "CreateVolume":
		v, err := svc.CreateVolume(context.Background(), createVolumeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating volume", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateVolume method", "response", v)

	case "DeleteVolume":
		v, err := svc.DeleteVolume(context.Background(), deleteVolumeReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error deleting volume", "error", err)
			os.Exit(1)
		}
		sugar.Infow("DeleteVolume method", "response", v)

	case "CreateSnapshot":
		v, err := svc.CreateSnapshot(context.Background(), createSnapshotReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating snapshot", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateSnapshot method", "response", v)

	case "CreateSnapshotAndDelete":
		v, err := svc.CreateSnapshotAndDelete(context.Background(), createSnapshotAndDeleteReq)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error creating snapshot and deleting volume", "error", err)
			os.Exit(1)
		}
		sugar.Infow("CreateSnapshotAndDelete method", "response", v)

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
