package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"go.uber.org/zap"
)

type AWSController struct {
	logger        *zap.SugaredLogger
	sessionClient func(region string, logger *zap.SugaredLogger) (awssession ec2iface.EC2API, err error)
}

func sessionClient(region string, logger *zap.SugaredLogger) (awsSession ec2iface.EC2API, err error) {

	accessKey, secretID, sessiontoken, stserr := GetCredsFromSTS(logger)

	if stserr != nil {
		logger.Errorw("Error getting Credentials", "error", stserr)
		return nil, stserr
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretID, sessiontoken),
	})

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}
	awsSvc := ec2.New(sess)

	return awsSvc, err
}

func NewAWSController(logger *zap.SugaredLogger) AWSController {

	return AWSController{logger, sessionClient}
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
