package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"go.uber.org/zap"
)

type AWSController struct {
	logger         *zap.SugaredLogger
	ec2SessFactory func(region string) (awssession ec2iface.EC2API, err error)
}

func Ec2SessionFactory(region string) (awsSession ec2iface.EC2API, err error) {
	sess, err := CreateBaseSession(region)
	if err != nil {
		return nil, errors.Wrap(err, "Can't start AWS session")
	}

	awsSvc := ec2.New(sess)
	return awsSvc, err
}

func NewAWSController(logger *zap.SugaredLogger) AWSController {
	return AWSController{logger, Ec2SessionFactory}
}

func (svc AWSController) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	return &pb.ClusterResponse{}, nil
}

func (svc AWSController) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {
	return &pb.ClusterSpec{}, nil
}

func (svc AWSController) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {
	return &pb.GetClustersResponse{}, nil
}

func (svc AWSController) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	return &pb.ClusterStatusResponse{}, nil
}

func (svc AWSController) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	return &pb.NodeSpawnResponse{}, nil
}

func (svc AWSController) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	return &pb.ClusterDeleteResponse{}, nil
}

func (svc AWSController) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	return &pb.NodeDeleteResponse{}, nil
}

func (svc AWSController) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	return &pb.AddTokenResponse{}, nil
}

func (svc AWSController) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	return &pb.GetTokenResponse{}, nil
}
