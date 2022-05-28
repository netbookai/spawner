package gcp

import (
	"context"
	"fmt"

	container_proto "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"

	"github.com/davecgh/go-spew/spew"
	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

//getParent retrive getParent path for clusters
func getParent(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}

func getClusterFQName(projectId, region, cluster string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectId, region, cluster)
}

func (g *GCPController) createCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
	}

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster")
	}
	defer client.Close()

	//cluster tags
	tags := make(map[string]string)
	for k, v := range labels.DefaultTags() {
		tags[k] = *v
	}

	for k, v := range req.Labels {
		tags[k] = v

	}

	np, err := getNodePool(req.Node)

	if err != nil {
		g.logger.Error(ctx, "failed to get the nodepool config from the spec", "error", err)
		return nil, errors.Wrap(err, "CreateCluster: getNodePool returned error")
	}

	cluster := &container_proto.Cluster{
		Name:           req.ClusterName,
		Description:    "Spawner managed cluster",
		NodePools:      []*container_proto.NodePool{np},
		ResourceLabels: tags,
		ReleaseChannel: &container_proto.ReleaseChannel{
			//to disable auto repair, we have to specify
			Channel: container_proto.ReleaseChannel_UNSPECIFIED,
		},
		//InitialClusterVersion: 1.21.10-gke.2000,

	}

	cr := &container_proto.CreateClusterRequest{
		Cluster: cluster,
		Parent:  getParent(cred.ProjectId, req.Region),
	}

	g.logger.Info(ctx, "creating cluster in gcp", "name", req.ClusterName, "region", req.Region)
	// Doc : https://pkg.go.dev/google.golang.org/genproto/googleapis/container/v1#UnimplementedClusterManagerServer.CreateCluster
	res, err := client.CreateCluster(ctx, cr)
	if err != nil {
		g.logger.Error(ctx, "failed to create cluster in gcp", "error", err)
		return nil, errors.Wrap(err, "createCluster")
	}

	if res.GetError() != nil {
		g.logger.Error(ctx, "failed to create cluster in gcp", "error", res.GetError().Message)
		return nil, errors.New(res.GetError().GetMessage())
	}
	g.logger.Info(ctx, "cluster created in gcp")

	return &proto.ClusterResponse{
		ClusterName: req.ClusterName,
	}, nil
}

func (g *GCPController) getCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}
	defer client.Close()

	name := getClusterFQName(cred.ProjectId, req.Region, req.ClusterName)
	cluster, err := client.GetCluster(ctx, &container_proto.GetClusterRequest{
		Name: name,
	})

	if err != nil {

		if e, ok := err.(*apierror.APIError); ok {
			st := e.GRPCStatus()
			g.logger.Error(ctx, "cluster not found", "cluster", req.ClusterName, "code", st.Code(), "reason", st.Details())
			if st.Code() == codes.NotFound {
				return nil, errors.New("cluster not found")
			}
			//TODO: might be some other rpc error
			return nil, errors.Wrap(st.Err(), "cluster not found")
		}
		return nil, errors.Wrap(err, "getCluster:")
	}
	g.logger.Info(ctx, "cluster found")

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

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "getClusters:")
	}
	defer client.Close()

	parent := getParent(cred.ProjectId, req.Region)
	g.logger.Info(ctx, "fetching clusters", "parent", parent)
	resp, err := client.ListClusters(ctx, &container_proto.ListClustersRequest{
		Parent: parent,
	})

	if err != nil {
		return nil, errors.Wrap(err, "getClusters:")
	}
	clusters := resp.Clusters
	//TODO: list cluster
	spew.Dump(clusters)
	return nil, nil
}

func (g *GCPController) clusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "clusterStatus:")
	}

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "clusterStatus:")
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", cred.ProjectId, req.Region, req.ClusterName)
	g.logger.Info(ctx, "fetching cluster", "name", name)
	cluster, err := client.GetCluster(ctx, &container_proto.GetClusterRequest{
		Name: name,
	})

	if err != nil {

		if e, ok := err.(*apierror.APIError); ok {
			st := e.GRPCStatus()
			g.logger.Error(ctx, "cluster not found", "cluster", req.ClusterName, "code", st.Code(), "reason", st.Details())
			if st.Code() == codes.NotFound {
				return nil, errors.New("cluster not found")
			}

			//TODO: might be some other rpc error
			return nil, errors.Wrap(st.Err(), "cluster not found")
		}

		g.logger.Error(ctx, "failed to get cluster", "error", err)
		return nil, errors.Wrap(err, "clusterStatus:")
	}
	g.logger.Info(ctx, "cluster status", "status", cluster.Status)
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

	client, err := getClusterManagerClient(ctx, cred)
	if err != nil {
		return nil, errors.Wrap(err, "deleteCluster:")
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", cred.ProjectId, req.Region, req.ClusterName)
	g.logger.Info(ctx, "deleting cluster", "name", name)
	res, err := client.DeleteCluster(ctx, &container_proto.DeleteClusterRequest{
		Name: name,
	})

	if err != nil {

		if e, ok := err.(*apierror.APIError); ok {
			st := e.GRPCStatus()
			g.logger.Error(ctx, "cluster not found", "cluster", req.ClusterName, "code", st.Code(), "reason", st.Details())

			if st.Code() == codes.NotFound {
				return nil, errors.New("cluster not found")
			}

			//TODO: might be some other rpc error
			return nil, errors.Wrap(st.Err(), "cluster not found")
		}
		return nil, errors.Wrap(err, "deleteCluster:")
	}

	if res.GetError() != nil {
		g.logger.Error(ctx, "failed to delete cluster in gcp", "error", res.GetError().Message)
		return nil, errors.New(res.GetError().GetMessage())
	}

	g.logger.Info(ctx, "cluster deleted in gcp")
	return &proto.ClusterDeleteResponse{}, nil
}
