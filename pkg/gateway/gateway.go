package gateway

import (
	"context"

	"gitlab.com/netbook-devs/spawner-service/pkg/service"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

type gateway struct {
	service service.SpawnerService

	proto.UnimplementedSpawnerServiceServer
}

func New(s service.SpawnerService) proto.SpawnerServiceServer {
	return &gateway{
		service: s,
	}
}

func (g *gateway) HealthCheck(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {

	return &proto.Empty{}, nil
}

func (g *gateway) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {

	return &proto.EchoResponse{
		Msg: req.Msg,
	}, nil
}

//CreateCluster Spawn required cluster
func (g *gateway) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	return g.service.CreateCluster(ctx, req)
}

//AddToken Create add token to secret manager --deprecated
func (g *gateway) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	return g.service.AddToken(ctx, req)
}

// GetToken Get kubernetes token
func (g *gateway) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return g.service.GetToken(ctx, req)
}

//AddRoute53Record Add Route53 record for Caddy
func (g *gateway) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	return g.service.AddRoute53Record(ctx, req)
}

// GetCluster describe given cluster
func (g *gateway) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return g.service.GetCluster(ctx, req)
}

// GetClusters Retrieve all clusters in the user account
func (g *gateway) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return g.service.GetClusters(ctx, req)
}

//AddNode add node to the cluster
func (g *gateway) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	return g.service.AddNode(ctx, req)
}

// ClusterStatus retrieve cluster status such as 'ACTIVE', 'CREATING', 'DELETING'
func (g *gateway) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return g.service.ClusterStatus(ctx, req)
}

// DeleteCluster delete the given cluster
func (g *gateway) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return g.service.DeleteCluster(ctx, req)
}

// DeleteNode delete attached node
func (g *gateway) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	return g.service.DeleteNode(ctx, req)
}

// CreateVolume create a volume on the provider
func (g *gateway) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	return g.service.CreateVolume(ctx, req)
}

// DeleteVolume deletes volumes
func (g *gateway) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	return g.service.DeleteVolume(ctx, req)
}

//CreateSnapshot create snapshot of backing volume
func (g *gateway) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return g.service.CreateSnapshot(ctx, req)
}

//CreateSnapshotAndDelete create a snapshot of backing volume and delete the volume
func (g *gateway) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return g.service.CreateSnapshotAndDelete(ctx, req)
}

//RegisterWithRancher register the cluster with rancher, provides monitoring dashboard for the cluster
func (g *gateway) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {
	return g.service.RegisterWithRancher(ctx, req)
}

//GetWorkspacesCost
func (g *gateway) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {
	return g.service.GetWorkspacesCost(ctx, req)
}

//WriteCredential save user account credential
func (g *gateway) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {
	return g.service.WriteCredential(ctx, req)
}

//ReadCredential read user credential
func (g *gateway) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {
	return g.service.ReadCredential(ctx, req)
}

//GetKubeConfig retrieve kube config for the cluster
func (g *gateway) GetKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {
	return g.service.GetKubeConfig(ctx, req)
}

//TagNodeInstance tag underlying vm instances for cluster nodes
func (g *gateway) TagNodeInstance(ctx context.Context, req *proto.TagNodeInstanceRequest) (*proto.TagNodeInstanceResponse, error) {
	return g.service.TagNodeInstance(ctx, req)
}
