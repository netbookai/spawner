package spawnerservice

import (
	"context"

	"github.com/go-kit/kit/metrics"
	"go.uber.org/zap"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	pb "gitlab.com/netbook-devs/spawner-service/pb"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher"

	"gitlab.com/netbook-devs/spawner-service/pkg/config"
)

type ClusterController interface {
	CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error)
	GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error)
	GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error)
	AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error)
	GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error)
	AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (*pb.AddRoute53RecordResponse, error)
	ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error)
	AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error)
	DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error)
	DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error)
	CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error)
	CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error)
	CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error)

	// Provider contoller need not to implement this
	RegisterWithRancher(ctx context.Context, req *pb.RancherRegistrationRequest) (*pb.RancherRegistrationResponse, error)
}

//SpawnerService manage provider and clusters
type SpawnerService struct {
	awsController  ClusterController
	noopController ClusterController
	logger         *zap.SugaredLogger
	config         *config.Config
}

var _ ClusterController = (*SpawnerService)(nil)

//New
func New(logger *zap.SugaredLogger, config *config.Config, ints metrics.Counter) ClusterController {

	var svc ClusterController
	svc = SpawnerService{
		awsController:  aws.NewAWSController(logger, config),
		noopController: &NoopController{},
		logger:         logger,
		config:         config,
	}
	svc = LoggingMiddleware(logger)(svc)
	svc = InstrumentingMiddleware(ints)(svc)
	return svc
}

func (svc SpawnerService) controller(provider string) ClusterController {
	switch provider {
	case "aws":
		return svc.awsController
	}
	return svc.noopController
}

//CreateCluster create cluster on the provider specified in request
func (svc SpawnerService) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	return svc.controller(req.Provider).CreateCluster(ctx, req)
}

//GetCluster get cluster on the providerr specified in request
func (svc SpawnerService) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {
	return svc.controller(req.Provider).GetCluster(ctx, req)
}

//GetClusters get the available clusters in the given provider
func (svc SpawnerService) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {
	return svc.controller(req.Provider).GetClusters(ctx, req)
}

//AddToken deprecated as of now
func (svc SpawnerService) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	return svc.controller(req.Provider).AddToken(ctx, req)
}

//GetToken return the kube token for the cluster in given provider
func (svc SpawnerService) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	return svc.controller(req.Provider).GetToken(ctx, req)
}

//AddRoute53Record
func (svc SpawnerService) AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (*pb.AddRoute53RecordResponse, error) {
	return svc.controller(req.Provider).AddRoute53Record(ctx, req)
}

//ClusterStatus get cluster status in given provider
func (svc SpawnerService) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	return svc.controller(req.Provider).ClusterStatus(ctx, req)
}

//AddNode adds new node to the cluster on the provider
func (svc SpawnerService) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	return svc.controller(req.Provider).AddNode(ctx, req)
}

//DeleteCluster deletes empty cluster on the provider, fails when cluster has nodegroup
func (svc SpawnerService) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	return svc.controller(req.Provider).DeleteCluster(ctx, req)
}

//DeleteNode deletes node on the given provider cluster
func (svc SpawnerService) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	return svc.controller(req.Provider).DeleteNode(ctx, req)
}

//CreateVolume create new volume on the provider
func (svc SpawnerService) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	return svc.controller(req.Provider).CreateVolume(ctx, req)
}

//DeleteVolume delete the volumne on the provider
func (svc SpawnerService) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	return svc.controller(req.Provider).DeleteVolume(ctx, req)
}

//CreateSnapshot
func (svc SpawnerService) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	return svc.controller(req.Provider).CreateSnapshot(ctx, req)
}

//CreateSnapshotAndDelete
func (svc SpawnerService) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	return svc.controller(req.Provider).CreateSnapshotAndDelete(ctx, req)
}

//RegisterWithRancher register cluster on the rancher, returns the kube manifest to apply on the cluster
func (svc SpawnerService) RegisterWithRancher(ctx context.Context, req *pb.RancherRegistrationRequest) (*pb.RancherRegistrationResponse, error) {

	clusterName := req.ClusterName
	svc.logger.Info("registering cluster with rancher ", req.ClusterName)

	client, err := rancher.CreateRancherClient(svc.config.RancherAddr, svc.config.RancherUsername, svc.config.RancherPassword)

	if err != nil {
		svc.logger.Error("failed to get rancher client ", client)

		return nil, err
	}

	regCluster := rnchrClient.Cluster{
		DockerRootDir:           "/var/lib/docker",
		Name:                    req.ClusterName,
		WindowsPreferedCluster:  false,
		EnableClusterAlerting:   false,
		EnableClusterMonitoring: false,
	}

	registeredCluster, err := client.Cluster.Create(&regCluster)

	if err != nil {
		svc.logger.Errorf("failed to create a rancher cluster '%s' %s", clusterName, err.Error())
		return nil, err
	}

	registrationToken, err := client.ClusterRegistrationToken.Create(&rnchrClient.ClusterRegistrationToken{
		ClusterID: registeredCluster.ID,
	})

	if err != nil {
		//TODO: we may want to revert the creation process,
		//but we will keep it now, so we can manually deal with the registration in case of failure.

		svc.logger.Errorf("failed to fetch registration token for '%s' %s", clusterName, err.Error())
		return nil, err
	}
	svc.logger.Infof("cluster created on the rancher, apply the manifest file on the target cluster '%s'", registrationToken.ManifestURL)

	return &pb.RancherRegistrationResponse{
		ClusterName: registeredCluster.Name,
		ClusterID:   registrationToken.ClusterID,
		ManifestURL: registrationToken.ManifestURL,
	}, nil

}
