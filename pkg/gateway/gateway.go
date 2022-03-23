package gateway

import (
	"context"

	"gitlab.com/netbook-devs/spawner-service/pkg/service"
	"gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

type gateway struct {
	service service.SpawnerService

	proto.UnimplementedSpawnerServiceServer
}

func New(s service.SpawnerService) spawnerservice.SpawnerServiceServer {
	return &gateway{
		service: s,
	}
}

func (g *gateway) HealthCheck(ctx context.Context, req *proto.Empty) (*spawnerservice.Empty, error) {

	return &proto.Empty{}, nil
}

func (g *gateway) Echo(ctx context.Context, req *proto.EchoRequest) (*spawnerservice.EchoResponse, error) {

	return &proto.EchoResponse{
		Msg: req.Msg,
	}, nil
}

// Spawn required cluster
func (g *gateway) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*spawnerservice.ClusterResponse, error) {
	return g.service.CreateCluster(ctx, req)
}

// Create add token to secret manager
func (g *gateway) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	return g.service.AddToken(ctx, req)
}

// Create get token to secret manager
func (g *gateway) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return g.service.GetToken(ctx, req)
}

// Add Route53 record for Caddy
func (g *gateway) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	return g.service.AddRoute53Record(ctx, req)
}

// Get Cluster
func (g *gateway) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return g.service.GetCluster(ctx, req)
}

// Get Clusters
func (g *gateway) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return g.service.GetClusters(ctx, req)
}

// Spawn required instance
func (g *gateway) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	return g.service.AddNode(ctx, req)
}

// Status of cluster
func (g *gateway) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return g.service.ClusterStatus(ctx, req)
}

// Delete Cluster
func (g *gateway) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return g.service.DeleteCluster(ctx, req)
}

// Delete Node
func (g *gateway) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	return g.service.DeleteNode(ctx, req)
}

// Create Volume
func (g *gateway) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	return g.service.CreateVolume(ctx, req)
}

// Delete Vol
func (g *gateway) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	return g.service.DeleteVolume(ctx, req)
}

func (g *gateway) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return g.service.CreateSnapshot(ctx, req)
}

func (g *gateway) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return g.service.CreateSnapshotAndDelete(ctx, req)
}

func (g *gateway) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {
	return g.service.RegisterWithRancher(ctx, req)
}

func (g *gateway) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {
	return g.service.GetWorkspacesCost(ctx, req)
}

func (g *gateway) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {
	return g.service.WriteCredential(ctx, req)
}

func (g *gateway) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {
	return g.service.ReadCredential(ctx, req)
}
