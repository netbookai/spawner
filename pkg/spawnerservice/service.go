package spawnerservice

import (
	"context"

	"github.com/go-kit/kit/metrics"
	"go.uber.org/zap"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	pb "gitlab.com/netbook-devs/spawner-service/pb"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher/common"

	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher"
	"gitlab.com/netbook-devs/spawner-service/pkg/util"
)

type ClusterController interface {
	CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error)
	GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error)
	GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error)
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

type SpawnerService struct {
	rancherController ClusterController
	awsController     ClusterController
}

func New(logger *zap.SugaredLogger, config *util.Config, ints metrics.Counter) ClusterController {

	rancherClient, err := common.CreateRancherClient(config.RancherAddr, config.RancherUsername, config.RancherPassword)
	if err != nil {
		logger.Errorw("error creating rancher client", "error", err)
	}
	spawnerServiceRancher := rancher.NewSpawnerServiceClient(rancherClient)

	var svc ClusterController

	//TODO: shivani use region variable here instead of harcoding us-west-2
	//starts an AWS session
	accessKey, secretID, sessiontoken, stserr := aws.GetCredsFromSTS()
	if stserr != nil {
		logger.Errorw("Error getting Credentials")
	}

	sess, err := session.NewSession(&awssdk.Config{
		Region:      awssdk.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(accessKey, secretID, sessiontoken),
	})

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
	}
	awsSvc := ec2.New(sess)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
	}

	{
		svc = SpawnerService{
			rancherController: rancher.NewRancherController(spawnerServiceRancher, config, logger),
			awsController:     aws.NewAWSController(logger, awsSvc),
		}
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(ints)(svc)
	}
	return svc
}

func (svc SpawnerService) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	return svc.rancherController.CreateCluster(ctx, req)
}

func (svc SpawnerService) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {
	return svc.rancherController.GetCluster(ctx, req)
}

func (svc SpawnerService) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {
	return svc.rancherController.GetClusters(ctx, req)
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
