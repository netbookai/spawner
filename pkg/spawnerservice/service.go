package spawnerservice

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	pb "gitlab.com/netbook-devs/spawner-service/pb"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"

	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher"
	"gitlab.com/netbook-devs/spawner-service/pkg/util"
)

type ClusterController interface {
	CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error)
	AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error)
	GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error)
	ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error)
	AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error)
	DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error)
	DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error)
	CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error)
	CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error)
	CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error)
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
	awsController     ClusterController
}

<<<<<<< HEAD
func New(logger log.Logger, config util.Config) ClusterController {
	rancherController, err := rancher.NewRancherController(logger, config)
	if err != nil {
		fmt.Println(fmt.Errorf("error creating rancher client %s", err))
		os.Exit(1)
	}
	return SpawnerService{
		rancherController: rancherController,
		awsController:     aws.AWSController{},
=======
func New(logger log.Logger, config util.Config, ints metrics.Counter) ClusterController {
	var svc ClusterController
	{
		svc = SpawnerService{
			rancherController: rancher.NewRancherController(logger, config),
			awsController:     aws.AWSController{},
		}
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(ints)(svc)
>>>>>>> 7984f17 (Implemented logging middleware for cluster controller)
	}
	return svc
}

func (svc SpawnerService) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	return svc.rancherController.CreateCluster(ctx, req)
}

func (svc SpawnerService) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	return svc.rancherController.AddToken(ctx, req)
}

func (svc SpawnerService) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	return svc.rancherController.GetToken(ctx, req)
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

func (svc SpawnerService) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	return svc.awsController.CreateSnapshot(ctx, req)
}

func (svc SpawnerService) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	return svc.awsController.CreateSnapshotAndDelete(ctx, req)
}
