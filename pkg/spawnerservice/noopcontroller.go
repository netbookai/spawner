package spawnerservice

import proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"

//NoopController No Op controller defaults to this controller call
//just for improved clarity on the controller and errors
type NoopController struct {
	proto.UnimplementedSpawnerServiceServer
}

var _ ClusterController = (*NoopController)(nil)

//var ProviderNotFoundError = errors.New("provider not found")
//
//func (ctrl *NoopController) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
//	return nil, ProviderNotFoundError
//}
//
//func (ctrl *NoopController) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {
//	return
//}
