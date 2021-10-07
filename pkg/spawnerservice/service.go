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
	CreateVol(ctx context.Context, req *pb.CreateVolReq) (*pb.CreateVolRes, error)
	DeleteVol(ctx context.Context, req *pb.DeleteVolReq) (*pb.DeleteVolRes, error)
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

func (svc SpawnerService) CreateVol(ctx context.Context, req *pb.CreateVolReq) (*pb.CreateVolRes, error) {
	return svc.awsController.CreateVol(ctx, req)
}

func (svc SpawnerService) DeleteVol(ctx context.Context, req *pb.DeleteVolReq) (*pb.DeleteVolRes, error) {
	return svc.awsController.DeleteVol(ctx, req)
}
