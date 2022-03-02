package spawnerservice

import (
	"context"

	"go.uber.org/zap"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher"

	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

type ClusterController interface {
	CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error)
	GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error)
	GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error)
	AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error)
	GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error)
	AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error)
	ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error)
	AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error)
	DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error)
	DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error)
	CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error)
	CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error)
	CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error)
	GetWorkspaceCost(context.Context, *proto.GetWorkspaceCostRequest) (*proto.GetWorkspaceCostResponse, error)

	// Provider contoller need not to implement this
	RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error)

	WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error)
	ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error)
}

//SpawnerService manage provider and clusters
type SpawnerService struct {
	awsController  ClusterController
	noopController ClusterController
	logger         *zap.SugaredLogger
	config         *config.Config
}

var _ ClusterController = (*SpawnerService)(nil)

//New return ClusterController
func New(logger *zap.SugaredLogger, config *config.Config) ClusterController {

	svc := &SpawnerService{
		awsController:  aws.NewAWSController(logger, config),
		noopController: &NoopController{},
		logger:         logger,
		config:         config,
	}
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
func (svc SpawnerService) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	return svc.controller(req.Provider).CreateCluster(ctx, req)
}

//GetCluster get cluster on the providerr specified in request
func (svc SpawnerService) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return svc.controller(req.Provider).GetCluster(ctx, req)
}

//GetClusters get the available clusters in the given provider
func (svc SpawnerService) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return svc.controller(req.Provider).GetClusters(ctx, req)
}

//AddToken deprecated as of now
func (svc SpawnerService) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	return svc.controller(req.Provider).AddToken(ctx, req)
}

//GetToken return the kube token for the cluster in given provider
func (svc SpawnerService) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return svc.controller(req.Provider).GetToken(ctx, req)
}

//AddRoute53Record
func (svc SpawnerService) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	return svc.controller(req.Provider).AddRoute53Record(ctx, req)
}

//ClusterStatus get cluster status in given provider
func (svc SpawnerService) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return svc.controller(req.Provider).ClusterStatus(ctx, req)
}

//AddNode adds new node to the cluster on the provider
func (svc SpawnerService) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	return svc.controller(req.Provider).AddNode(ctx, req)
}

//DeleteCluster deletes empty cluster on the provider, fails when cluster has nodegroup
func (svc SpawnerService) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return svc.controller(req.Provider).DeleteCluster(ctx, req)
}

//DeleteNode deletes node on the given provider cluster
func (svc SpawnerService) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	return svc.controller(req.Provider).DeleteNode(ctx, req)
}

//CreateVolume create new volume on the provider
func (svc SpawnerService) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	return svc.controller(req.Provider).CreateVolume(ctx, req)
}

//DeleteVolume delete the volumne on the provider
func (svc SpawnerService) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	return svc.controller(req.Provider).DeleteVolume(ctx, req)
}

//CreateSnapshot
func (svc SpawnerService) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return svc.controller(req.Provider).CreateSnapshot(ctx, req)
}

//CreateSnapshotAndDelete
func (svc SpawnerService) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return svc.controller(req.Provider).CreateSnapshotAndDelete(ctx, req)
}

//GetWorkspaceCost returns workspace cost grouped by given group
func (svc SpawnerService) GetWorkspaceCost(ctx context.Context, req *proto.GetWorkspaceCostRequest) (*proto.GetWorkspaceCostResponse, error) {
	return svc.controller(req.Provider).GetWorkspaceCost(ctx, req)
}

//RegisterWithRancher register cluster on the rancher, returns the kube manifest to apply on the cluster
func (svc SpawnerService) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {

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

	return &proto.RancherRegistrationResponse{
		ClusterName: registeredCluster.Name,
		ClusterID:   registrationToken.ClusterID,
		ManifestURL: registrationToken.ManifestURL,
	}, nil

}

//WriteCredential
func (svc SpawnerService) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {

	region := svc.config.SecretHostRegion
	id := req.GetAccessKeyID()
	key := req.GetSecretAccessKey()
	account := req.GetAccount()

	err := svc.writeCredentials(ctx, region, account, id, key)
	if err != nil {
		svc.logger.Errorw("failed to save credentials", "error", err, "account", account)
		return nil, err
	}
	return &proto.WriteCredentialResponse{}, nil

}

//ReadCredential
func (svc SpawnerService) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {

	region := svc.config.SecretHostRegion
	account := req.GetAccount()

	creds, err := svc.getCredentials(ctx, region, account)
	if err != nil {
		svc.logger.Errorw("failed to get the credentials", "account", account)
		return nil, err
	}
	svc.logger.Debugw("credentials found", "account", account, "accessKeyID", creds.AccessKeyID)
	return &proto.ReadCredentialResponse{
		Account:         account,
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
	}, nil
}
