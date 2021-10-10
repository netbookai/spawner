package spawnerservice

import (
	"context"

	"github.com/go-kit/kit/log"

	pb "gitlab.com/netbook-devs/spawner-service/pb"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"

	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher"
	"gitlab.com/netbook-devs/spawner-service/pkg/util"
)

type ClusterController interface {
	CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error)
	ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error)
	AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error)
	DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error)
	DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error)
	CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error)
	CreateSnapshot(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error)
	CreateSnapshotAndDelete(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error)
}

// func New(logger log.Logger, config util.Config) ClusterController {
// 	var svc ClusterController
// 	{
// 		// TODO: Sid pass logger to impls and log
// 		svc = rancher.NewRancherController(logger, config)
// 	}
// 	return svc
// }

type SpawnerService struct {
	rancherController ClusterController
	awsController     aws.AWSController
}

func New(logger log.Logger, config util.Config) ClusterController {
	return SpawnerService{
		rancherController: rancher.NewRancherController(logger, config),
		awsController:     aws.AWSController{},
	}
}

func (svc SpawnerService) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	return svc.rancherController.CreateCluster(ctx, req)
}

func (svc SpawnerService) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	return svc.rancherController.ClusterStatus(ctx, req)
}

func (svc SpawnerService) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	return svc.rancherController.AddNode(ctx, req)
}

func (svc SpawnerService) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	return svc.rancherController.DeleteCluster(ctx, req)
}

func (svc SpawnerService) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	return svc.rancherController.DeleteNode(ctx, req)
}

func (svc SpawnerService) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	return svc.awsController.CreateVolume(ctx, req)
}

func (svc SpawnerService) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	return svc.awsController.DeleteVolume(ctx, req)
}

func (svc SpawnerService) CreateSnapshot(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {
	return svc.awsController.CreateSnapshot(ctx, req)
}

func (svc SpawnerService) CreateSnapshotAndDelete(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {
	return svc.awsController.CreateSnapshotAndDelete(ctx, req)
}
