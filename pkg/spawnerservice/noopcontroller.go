package spawnerservice

import (
	pb "gitlab.com/netbook-devs/spawner-service/pb"
)

//NoopController No Op controller defaults to this controller call
//just for improved clarity on the controller and errors
type NoopController struct {
	pb.UnimplementedSpawnerServiceServer
}

var _ ClusterController = (*NoopController)(nil)

//var ProviderNotFoundError = errors.New("provider not found")
//
//func (ctrl *NoopController) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (*pb.AddRoute53RecordResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) RegisterWithRancher(ctx context.Context, req *pb.RancherRegistrationRequest) (*pb.RancherRegistrationResponse, error) {
//	return
//}
