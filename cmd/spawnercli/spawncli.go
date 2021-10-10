package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"

	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwntransport"
)

func main() {
	fs := flag.NewFlagSet("spawncli", flag.ExitOnError)
	var (
		// httpAddr = fs.String("http-addr", "", "HTTP address of addsvc")
		grpcAddr = fs.String("grpc-addr", "", "gRPC address of addsvc")
		method   = fs.String("method", "CreateCluster", "CreateCluster")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags] <a> <b>")
	fs.Parse(os.Args[1:])

	// This is a demonstration client, which supports multiple transports.
	var (
		svc spawnerservice.ClusterController
		err error
	)
	// if *httpAddr != "" {
	// 	// svc, err = spwntransport.NewHTTPClient(*httpAddr, log.NewNopLogger())
	// } else if *grpcAddr != "" {
	if *grpcAddr != "" {
		conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
			os.Exit(1)
		}
		defer conn.Close()
		svc = spwntransport.NewGRPCClient(conn, log.NewNopLogger())
	} else {
		fmt.Fprintf(os.Stderr, "error: no remote address specified\n")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
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

	clusterStatusReq := &pb.ClusterStatusRequest{
		ClusterName: "aws-us-west-2-eks-7",
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
	}

	deleteVolumeReq := &pb.DeleteVolumeRequest{
		Volumeid: "vol-0f92bc5d6328b2559",
	}

	snapshotReq := &pb.SnapshotRequest{
		Volumeid: "vol-0187ea5c851bfee54",
	}

	switch *method {
	case "CreateCluster":
		v, err := svc.CreateCluster(context.Background(), createClusterReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)
	case "ClusterStatus":
		v, err := svc.ClusterStatus(context.Background(), clusterStatusReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)
	case "AddNode":
		v, err := svc.AddNode(context.Background(), addNodeReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)
	case "DeleteCluster":
		v, err := svc.DeleteCluster(context.Background(), deleteClusterReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)
	case "DeleteNode":
		v, err := svc.DeleteNode(context.Background(), deleteNodeReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)

	case "CreateVolume":
		v, err := svc.CreateVolume(context.Background(), createVolumeReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)

	case "DeleteVolume":
		v, err := svc.DeleteVolume(context.Background(), deleteVolumeReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)

	case "CreateSnapshot":
		v, err := svc.CreateSnapshot(context.Background(), snapshotReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)

	case "CreateSnapshotAndDelete":
		v, err := svc.CreateSnapshotAndDelete(context.Background(), snapshotReq)
		if err != nil && err.Error() != "" {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%v", v)

	default:
		fmt.Fprintf(os.Stderr, "error: invalid method %q\n", *method)
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
