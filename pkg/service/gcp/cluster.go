package gcp

import (
	"context"
	"fmt"

	container_proto "google.golang.org/genproto/googleapis/container/v1"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

//getParent retrive getParent path for clusters
func getParent(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}

func getDiskType() string {
	// Doc : https://cloud.google.com/compute/docs/disks
	return "pd-standard"
}

func (g *GCPController) createCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
	}

	client, err := getClusterClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster")
	}
	defer client.Close()

	node := req.Node
	nodeCount := int32(1)
	if node.Count != 0 {
		nodeCount = int32(node.Count)
	}
	//node labels
	label := make(map[string]string)
	for k, v := range labels.GetNodeLabel(node) {
		label[k] = *v
	}

	//cluster tags
	tags := make(map[string]string)
	for k, v := range labels.DefaultTags() {
		tags[k] = *v
	}

	for k, v := range req.Labels {
		tags[k] = v

	}

	instance := ""
	if node.MachineType != "" {
		instance = common.GetInstance(constants.GcpLabel, node.MachineType)
	} else {
		instance = node.Instance
	}

	if instance == "" {
		return nil, errors.New(constants.InvalidInstanceOrMachineType)
	}

	//Doc : https://cloud.google.com/kubernetes-engine/docs/concepts/node-images#available_node_images
	imageType := "COS_CONTAINERD"
	diskType := getDiskType()

	nodeConfig := &container_proto.NodeConfig{

		MachineType: instance,
		DiskSizeGb:  node.DiskSize,
		ImageType:   imageType,
		Preemptible: false,
		DiskType:    diskType,
	}

	cluster := &container_proto.Cluster{
		Name:        req.ClusterName,
		Description: "Spawner managed cluster",
		NodePools: []*container_proto.NodePool{
			{

				Name:             node.Name,
				Config:           nodeConfig,
				InitialNodeCount: nodeCount,
				//in case we use Zonal
				//Locations:        []string{},
				Autoscaling: &container_proto.NodePoolAutoscaling{
					Enabled:      true,
					MinNodeCount: nodeCount,
					MaxNodeCount: nodeCount,
				},
			},
		},
		ResourceLabels: tags,
		//InitialClusterVersion: 1.21.10-gke.2000,
	}

	cr := &container_proto.CreateClusterRequest{
		Cluster: cluster,
		Parent:  getParent(cred.ProjectId, req.Region),
	}

	g.logger.Infow("creating cluster in gcp", "name", req.ClusterName, "region", req.Region)
	res, err := client.CreateCluster(ctx, cr)
	if err != nil {
		g.logger.Errorw("failed to create cluster in gcp", "error", err)
		return nil, errors.Wrap(err, "createCluster")
	}

	if res.GetError() != nil {
		g.logger.Errorw("failed to create cluster in gcp", "error", res.GetError().Message)
		return nil, errors.New(res.GetError().GetMessage())
	}
	g.logger.Infow("cluster created in gcp")

	return &proto.ClusterResponse{
		ClusterName: req.ClusterName,
	}, nil
}

func (g *GCPController) getCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}

	client, err := getClusterClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", cred.ProjectId, req.Region, req.ClusterName)
	cluster, err := client.GetCluster(ctx, &container_proto.GetClusterRequest{
		Name: name,
	})

	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}

	g.logger.Infow("cluster found")
	spew.Dump(cluster)
	return &proto.ClusterSpec{
		Name:      cluster.GetName(),
		ClusterId: cluster.Id,
		NodeSpec:  []*proto.NodeSpec{},
	}, nil
}

func (g *GCPController) getClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "getClusters:")
	}

	client, err := getClusterClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "getClusters:")
	}
	defer client.Close()

	parent := getParent(cred.ProjectId, req.Region)
	g.logger.Infow("fetching clusters", "parent", parent)
	resp, err := client.ListClusters(ctx, &container_proto.ListClustersRequest{
		Parent: parent,
	})

	if err != nil {
		return nil, errors.Wrap(err, "getClusters:")
	}
	clusters := resp.Clusters
	spew.Dump(clusters)
	return nil, nil
}

func (g *GCPController) clusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "clusterStatus:")
	}

	client, err := getClusterClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "clusterStatus:")
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", cred.ProjectId, req.Region, req.ClusterName)
	g.logger.Infow("fetching cluster", "name", name)
	cluster, err := client.GetCluster(ctx, &container_proto.GetClusterRequest{
		Name: name,
	})

	if err != nil {
		return nil, errors.Wrap(err, "clusterStatus:")
	}
	g.logger.Infow("cluster status", "status", cluster.Status)
	stat := constants.Inactive
	if cluster.Status == container_proto.Cluster_RUNNING {
		stat = constants.Active
	}
	return &proto.ClusterStatusResponse{
		Status: stat,
	}, nil
}

func (g *GCPController) deleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "deleteCluster:")
	}

	client, err := getClusterClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "deleteCluster:")
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", cred.ProjectId, req.Region, req.ClusterName)
	g.logger.Infow("deleting cluster", "name", name)
	res, err := client.DeleteCluster(ctx, &container_proto.DeleteClusterRequest{
		Name: name,
	})

	if err != nil {
		return nil, errors.Wrap(err, "deleteCluster:")
	}

	if res.GetError() != nil {
		g.logger.Errorw("failed to delete cluster in gcp", "error", res.GetError().Message)
		return nil, errors.New(res.GetError().GetMessage())
	}
	g.logger.Infow("cluster deleted in gcp")
	return &proto.ClusterDeleteResponse{}, nil
}
