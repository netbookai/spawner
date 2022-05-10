package gcp

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"go.uber.org/zap"
)

type GCPController struct {
	logger *zap.SugaredLogger
}

func NewController(logger *zap.SugaredLogger) *GCPController {
	return &GCPController{
		logger: logger,
	}
}

func (g *GCPController) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	return g.createCluster(ctx, req)
}

func (g *GCPController) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return g.getCluster(ctx, req)
}

func (g *GCPController) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return g.getClusters(ctx, req)
}

func (g *GCPController) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return g.deleteCluster(ctx, req)
}

func (g *GCPController) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return g.clusterStatus(ctx, req)
}

func (g *GCPController) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	return &proto.AddTokenResponse{}, nil
}

func (g *GCPController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return g.getToken(ctx, req)
}

func (g *GCPController) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	return g.createVolume(ctx, req)
}

func (g *GCPController) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	return g.deleteVolume(ctx, req)
}

func (g *GCPController) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return g.createSnapshot(ctx, req)
}

func (g *GCPController) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return g.createSnapshotAndDelete(ctx, req)
}

func (g *GCPController) GetKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {
	return g.getKubeConfig(ctx, req)
}

func (g *GCPController) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {
	return nil, nil
}

func (g *GCPController) TagNodeInstance(ctx context.Context, req *proto.TagNodeInstanceRequest) (*proto.TagNodeInstanceResponse, error) {
	return &proto.TagNodeInstanceResponse{}, nil
}
