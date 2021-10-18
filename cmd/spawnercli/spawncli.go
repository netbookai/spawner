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
		Name:     "ng-01",
		Instance: "t3.medium",
		DiskSize: 14,
	}
	createClusterReq := &pb.ClusterRequest{
		Provider: "aws",
		Region:   "us-west-2",
		Node:     node,
		Labels:   map[string]string{},
	}

	addTokenReq := &pb.AddTokenRequest{
		ClusterName: "aws-us-west-2-eks-4",
		Region:      "us-west-2",
	}

	getTokenReq := &pb.GetTokenRequest{
		ClusterName: "infra-test",
		Region:      "us-west-2",
	}

	clusterStatusReq := &pb.ClusterStatusRequest{
		ClusterName: "infra-test",
	}

	addNode := &pb.NodeSpec{
		Name:     "ng-04",
		Instance: "t2.medium",
		DiskSize: 12,
	}

	addNodeReq := &pb.NodeSpawnRequest{
		ClusterName: "aws-us-west-2-eks-5",
		NodeSpec:    addNode,
	}

	deleteClusterReq := &pb.ClusterDeleteRequest{
		ClusterName: "aws-us-west-2-eks-7",
	}

	deleteNodeReq := &pb.NodeDeleteRequest{
		ClusterName:   "aws-us-west-2-eks-5",
		NodeGroupName: "ng-sid-01",
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
		Volumeid: "vol-033eb2271582ebf28",
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
