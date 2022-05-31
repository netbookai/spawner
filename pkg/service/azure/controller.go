package azure

import (
	"context"

	"github.com/netbookai/log"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

type AzureController struct {
	logger log.Logger
}

func NewController(logger log.Logger) *AzureController {
	return &AzureController{
		logger: logger,
	}
}

func (a *AzureController) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	return a.createCluster(ctx, req)
}

func (a *AzureController) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return a.getCluster(ctx, req)
}

func (a *AzureController) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return a.getClusters(ctx, req)
}

func (a *AzureController) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return a.clusterStatus(ctx, req)
}

func (a *AzureController) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	return a.addNode(ctx, req)
}

func (a *AzureController) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return a.deleteCluster(ctx, req)
}

func (a *AzureController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	return a.deleteNode(ctx, req)
}

func (a *AzureController) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	return &proto.AddTokenResponse{}, nil
}

func (a *AzureController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return a.getToken(ctx, req)
}

func (a *AzureController) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	return a.createVolume(ctx, req)
}

func (a *AzureController) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	return a.deleteVolume(ctx, req)
}

func (a *AzureController) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return a.createSnapshot(ctx, req)
}

func (a *AzureController) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return a.createSnapshotAndDelete(ctx, req)
}

func (a *AzureController) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {
	return a.getWorkspacesCost(ctx, req)
}

func (a *AzureController) GetApplicationsCost(ctx context.Context, req *proto.GetApplicationsCostRequest) (*proto.GetApplicationsCostResponse, error) {
	return a.getApplicationsCost(ctx, req)
}

func (a *AzureController) GetKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {
	return a.getKubeConfig(ctx, req)
}

func (a *AzureController) TagNodeInstance(ctx context.Context, req *proto.TagNodeInstanceRequest) (*proto.TagNodeInstanceResponse, error) {
	return &proto.TagNodeInstanceResponse{}, nil
}

func (a *AzureController) GetCostByTime(ctx context.Context, req *proto.GetCostByTimeRequest) (*proto.GetCostByTimeResponse, error) {
	return a.getCostByTime(ctx, req)
}

func (a *AzureController) GetContainerRegistryAuth(ctx context.Context, in *proto.GetContainerRegistryAuthRequest) (*proto.GetContainerRegistryAuthResponse, error) {
	return &proto.GetContainerRegistryAuthResponse{}, nil
}
