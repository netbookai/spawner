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

var sa_cred = `
{
  "type": "service_account",
  "project_id": "netbook-testing",
  "private_key_id": "b486bead29c9cb519c3d45b096b4eab99f9cc8d9",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCd9683jtEG9aHU\nlYUu4CYtdClThmHb0lJHRMNLtccb0BLQJVtmZCNuMMFPfJOgYTj3xiC20KvfRW9v\nsOgDPGltggRtuq0+OCNKH243PHlyZzSL4BS/BuxeRWGG2NneKqs7vD3ywkR7X2Nf\ng4ZpXS8yo9uxpzJopMiAPDJFAyh5W07oIe9gI/rzptyiMr/2MSPs1dHczxE/OZB8\n2NWdR0+1iSm9vLHHvSU3wGGl6fibZB91E0ENQ+9QllAJXqfnC4yLjTbYHnEfvLQa\n8UlpRZrafUMtsWWzC4dr+qT+MSMY5HX9agXFhTjMzN8bppLAw5HXtorKw1AoeX9/\nbq9+93SNAgMBAAECggEAFKAba6ilGECIMcaYDifMNFEfeD1ql5YdkhqjWUZRygrf\n+fd2uKbIjYGmK+e7Ksym8IsZCGW0m0ForG+vy4Rey6KXS3B9YEtaKDp0XJfzz4E0\nNjM64jpYMHLkqgO0ZrKxiuooOIMvB+DLi9QTf7xgBj+o1sha55jkaQHzGlmwNjAG\nFy6/dsP/Hc3ueXY2rUc+kh//3C/OebPHfzSZirsee00X5vTQWdz1YKWdSE3STl6u\nklaJCbriMAdaE0UibLiszz34AVx371pMQ+2NW13e70A/XN4LM1yuH23WLq5ON4S1\ndn0H/5dw+e/C99+4cVHh5ge+k+h3kbGfh0ZnzK0nDQKBgQDd/NFM6gEOBXV9NJQH\n1UoC/++RPGc2YAPvjKNhagmgQOJ1Bd5eS2b/S4aCvgfkfktPmp6ksLiGLeDZ/TJq\nTo4tKdm4orh9y322p+3Kf1jaeYa1meTmM80YeBpP86qpw4dzmlcFOuK8CbiTb3iu\nVBpV0ijzD3+ZpIU/FhI/W1N2xwKBgQC2K8LWprzhzqisbEDOOPVfpnzTRjc/+erz\nJphWVURKac3U0VYazXZU53/ZSKleF738TzfthLm5hndKWZ8pOfqW4VV01a3zsYU3\ne0a19qmjeatLGSghwT6RiJCsvgrzGdxOjMOHJkeka7MRb14E8WxRHaWgt42u+Lbq\nQOpNZzHWCwKBgQDD8Xzt3z+/GKJ0OgzQPTxvGWplUGPqYyYWNJWiTu7gPWWm1d9K\nbFQl1IyOqx5cWf4v7dNKm5LFHYnz4MK3g0+MHfzINRmUMCJvMBt9Ops7fTmi4oxh\nhifrCVhwaiyiXK0bJYjaXPf18r6xpRtpBWOZjUAIDA4dmFLlNJ42vm4V0QKBgHP0\nKOGeYi3M8BpIEXvyT2UhwORuFi7XshAxKdgSEBTZgdWLpaYLz909OWih0oR80kYu\nWmgKCnmnuHiP0TpZmEK/jTh/5mhuP2BQTHL4XYQbpsd3bM8HhP73kTcTBD82377z\n5GU7HXDvyJw5afv1e7+qAknpa/rKfwteZIT+QX9/AoGBALqKU4OdsD/RPr8wEy0C\ns0ME+xfOSvH3MGJghjQtuU4zmr05BOQTthnGYrJRhvkwmzNfCB6Yfb4JnTPj+JHU\nnZoDV2LtV4F+74j0ypS2LnOWvhuPnzxLFj9nhsWjckyGiLALv/7MnEQ5g8+wJ3+O\nLbcp60wqMBK5FNLwMEncSAxE\n-----END PRIVATE KEY-----\n",
  "client_email": "spawner-dev@netbook-testing.iam.gserviceaccount.com",
  "client_id": "116364912980375247261",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/spawner-dev%40netbook-testing.iam.gserviceaccount.com"
}
`

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
