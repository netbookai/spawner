package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	clusterName = "cluster-node-with-machinenode"
	region      = "us-east-2" //"eastus2" //"us-west-2"
	provider    = "aws"
	accountName = "netbook-aws"
	nodeName    = "rootnode"
	instance    = "Standard_A2_v2"
	volumeName  = "vol-20-20220404123522"
)

func main() {

	var logger, _ = zap.NewDevelopment()
	var sugar = logger.Sugar()
	defer sugar.Sync()

	fs := flag.NewFlagSet("testclient", flag.ExitOnError)
	grpcAddr := fs.String("grpc-addr", ":8071", "gRPC address of spawner")
	method := fs.String("method", "GetCostByTime", "default HealthCheck")
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
		Name:     nodeName,
		Instance: instance,
		DiskSize: 30,
	}
	createClusterReq := &proto.ClusterRequest{
		Provider: provider,
		Region:   region,
		Node:     node,
		Labels: map[string]string{
			"user": "dev-tester",
		},
		ClusterName: clusterName,
		AccountName: accountName,
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
		AccountName: accountName,
	}

	addRoute53RecordReq := &proto.AddRoute53RecordRequest{
		DnsName:    "20.85.85.202",
		RecordName: "*.1117907260.eastus2.azure.app.dev.netbook.ai",
		// Region:      region,
		Provider:    provider,
		AccountName: accountName,
		// RegionIdentifier: "Oregon region",
	}

	clusterStatusReq := &proto.ClusterStatusRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	getClustersReq := &proto.GetClustersRequest{
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	getClusterReq := &proto.GetClusterRequest{
		ClusterName: clusterName,
		Provider:    provider,
		Region:      region,
		AccountName: accountName,
	}

	addNode := &proto.NodeSpec{
		Name:          nodeName,
		Instance:      instance,
		MigProfile:    proto.MIGProfile_MIG3g,
		CapacityType:  proto.CapacityType_ONDEMAND,
		MachineType:   "m",
		SpotInstances: []string{"t2.small", "t3.small"},
		DiskSize:      20,
		GpuEnabled:    false,
		Labels: map[string]string{"cluster-name": clusterName,
			"node-name":   nodeName,
			"user":        "dev-tester",
			"workspaceid": "dev-tester",
		},
	}

	addNodeReq := &proto.NodeSpawnRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		NodeSpec:    addNode,
		AccountName: accountName,
	}

	deleteClusterReq := &proto.ClusterDeleteRequest{
		ClusterName: clusterName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
		ForceDelete: true,
	}

	deleteNodeReq := &proto.NodeDeleteRequest{
		ClusterName:   clusterName,
		NodeGroupName: nodeName,
		Region:        region,
		Provider:      provider,
		AccountName:   accountName,
	}

	createVolumeReq := &proto.CreateVolumeRequest{
		Availabilityzone: region,
		Volumetype:       "gp2",
		Size:             50,
		Snapshotid:       "vol-30-20220409151829-snapshot",
		SnapshotUri:      "snapshot-uri",
		Region:           region,
		Provider:         provider,
		AccountName:      accountName,
	}

	deleteVolumeReq := &proto.DeleteVolumeRequest{
		//		Volumeid:    "vol-eastus2-1-20220323121600",
		Volumeid:    volumeName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	createSnapshotReq := &proto.CreateSnapshotRequest{
		Volumeid:    volumeName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}
	createSnapshotAndDeleteReq := &proto.CreateSnapshotAndDeleteRequest{
		Volumeid:    volumeName,
		Region:      region,
		Provider:    provider,
		AccountName: accountName,
	}

	getWorkspacesCost := &proto.GetWorkspacesCostRequest{
		WorkspaceIds: []string{"d1411352-c14a-4a78-a1d6-44d4c199ba3a", "18638c97-7352-426e-a79e-241956188fed", "dceaf501-1775-4339-ba7b-ec6d98569d11"},
		Provider:     "aws",
		AccountName:  "netbook-aws-dev",
		StartDate:    "2022-04-01",
		EndDate:      "2022-05-01",
		Granularity:  "DAILY",
		CostType:     "BlendedCost",
		GroupBy: &proto.GroupBy{
			Type: "TAG",
			Key:  "workspaceid",
		},
	}

	getCostByTime := &proto.GetCostByTimeRequest{
		Ids:         []string{"d1411352-c14a-4a78-a1d6-44d4c199ba3a", "18638c97-7352-426e-a79e-241956188fed", "dceaf501-1775-4339-ba7b-ec6d98569d11"},
		Provider:    "aws",
		AccountName: "netbook-aws-dev",
		StartDate:   "2022-04-01",
		EndDate:     "2022-05-01",
		Granularity: "DAILY",
		GroupBy: &proto.GroupBy{
			Type: "TAG",
			Key:  "workspaceid",
		},
	}

	switch *method {
	case "Echo":
		v, err := client.Echo(context.Background(), &proto.EchoRequest{Msg: "hello spawner"})

		if err != nil && err.Error() != "" {
			sugar.Errorw("Echo", "error", err)
			os.Exit(1)
		}
		sugar.Infow("Echo", "response", v)

	case "HealthCheck":
		v, err := client.HealthCheck(context.Background(), &proto.Empty{})

		if err != nil && err.Error() != "" {
			sugar.Errorw("HealthCheck", "error", err)
			os.Exit(1)
		}
		sugar.Infow("HealthCheck", "response", v)

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
	case "GetWorkspacesCost":
		v, err := client.GetWorkspacesCost(context.Background(), getWorkspacesCost)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error getting workspaces cost", "error", err)
			os.Exit(1)
		}
		sugar.Infow("GetWorkspaceCost method", "response", v)

	case "ReadCredentialAws":
		v, err := client.ReadCredential(context.Background(), &proto.ReadCredentialRequest{
			Account: "alexis",
			Type:    "aws",
		})
		if err != nil {
			sugar.Errorw("error reading credentials", "error", err)
		}
		sugar.Infow("ReadCredential", "response", v)

	case "WriteCredentialAws":
		v, err := client.WriteCredential(context.Background(), &proto.WriteCredentialRequest{
			Account: "alexis",
			Type:    "aws",
			Cred: &proto.WriteCredentialRequest_AwsCred{
				AwsCred: &proto.AwsCredentials{
					AccessKeyID:     "access_id",
					SecretAccessKey: "secret_key",
					Token:           "token",
				},
			},
		})
		if err != nil {
			sugar.Errorw("error writing credentials", "error", err)
		}
		sugar.Infow("WriteCredentialAws", "response", v)
	case "ReadCredentialAzure":
		v, err := client.ReadCredential(context.Background(), &proto.ReadCredentialRequest{
			Account: "netbook-azure-dev",
			Type:    "azure",
		})
		if err != nil {
			sugar.Errorw("error reading credentials", "error", err)
		}
		sugar.Infow("ReadCredential", "response", v)

	case "WriteCredentialAzure":
		v, err := client.WriteCredential(context.Background(), &proto.WriteCredentialRequest{
			Account: "alex",
			Type:    "azure",
			Cred: &proto.WriteCredentialRequest_AzureCred{
				AzureCred: &proto.AzureCredentials{
					SubscriptionID: "subscription",
					TenantID:       "tenant_id",
					ClientID:       "client_id",
					ClientSecret:   "client_secret",
					ResourceGroup:  "resource_group",
				},
			},
		})
		if err != nil {
			sugar.Errorw("error writing credentials", "error", err)
			return
		}
		sugar.Infow("WriteCredentialAws", "response", v)

	case "ReadCredentialGitPAT":
		v, err := client.ReadCredential(context.Background(), &proto.ReadCredentialRequest{
			Account: "nsp-dev",
			Type:    "git-pat",
		})

		if err != nil {
			sugar.Errorw("error reading Git PAT ", err)
			return
		}
		sugar.Infow("ReadCredentialResponse_GitPat", "response", v)
	case "WriteCredentialGitPAT":
		v, err := client.WriteCredential(context.Background(), &proto.WriteCredentialRequest{
			Account: "nsp-dev",
			Type:    "git-pat",
			Cred: &proto.WriteCredentialRequest_GitPat{
				GitPat: &proto.GithubPersonalAccessToken{
					Token: "this-is-very-secret-token-thats-why-you-see-this-message-when-reading",
				},
			},
		})

		if err != nil {
			sugar.Errorw("error writing Git PAT ", err)
			return
		}
		sugar.Infow("WriteCredentialResponse_GitPat", "response", v)
	case "AddTag":
		v, err := client.TagNodeInstance(context.Background(), &proto.TagNodeInstanceRequest{
			Provider:    provider,
			Region:      region,
			AccountName: accountName,
			ClusterName: clusterName,
			Labels: map[string]string{
				"label1": "valuelabel1",
			},
		})

		if err != nil {
			sugar.Errorw("error adding tags to node", "error", err)
		}
		sugar.Infow("TagNodeInstane", "response", v)

	case "GetCostByTime":
		v, err := client.GetCostByTime(context.Background(), getCostByTime)
		if err != nil && err.Error() != "" {
			sugar.Errorw("error getting cost by time", "error", err)
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
