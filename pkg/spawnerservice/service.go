package spawnerservice

import (
	"context"

	"github.com/go-kit/kit/log"

	pb "gitlab.com/netbook-devs/spawner-service/pb"

	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher"
)

type ClusterController interface {
	CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error)
	ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error)
	AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error)
	DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error)
	DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error)
}

func New(logger log.Logger) ClusterController {
	var svc ClusterController
	{
		// TODO: Sid pass logger to impls and log
		svc = rancher.NewRancherController()
	}
	return svc
}
