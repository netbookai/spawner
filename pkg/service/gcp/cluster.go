package gcp

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	container_proto "google.golang.org/genproto/googleapis/container/v1"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"google.golang.org/api/option"
)

var sa_cred = ` `

func parent(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}

func (g *GCPController) createCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {

	cred := []byte(sa_cred)
	c, err := container.NewClusterManagerClient(ctx, option.WithCredentialsJSON(cred))
	if err != nil {
		return nil, errors.Wrap(err, "createCluster")
	}
	defer c.Close()

	node := req.Node
	nodeCount := int32(1)
	if node.Count != 0 {
		nodeCount = int32(node.Count)
	}
	diskType := "pd-standard"
	imageType := "COS_CONTAINERD"
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

	nodeConfig := &container_proto.NodeConfig{
		MachineType: node.Instance,
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

	projectId := ""
	cr := &container_proto.CreateClusterRequest{
		Cluster: cluster,
		Parent:  parent(projectId, req.Region),
	}
	_, err = c.CreateCluster(ctx, cr)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster")
	}

	return &proto.ClusterResponse{}, nil
}

func (g *GCPController) getCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return nil, nil
}

func (g *GCPController) getClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return nil, nil
}

func (g *GCPController) clusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return nil, nil
}
