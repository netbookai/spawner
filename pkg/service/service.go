package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/service/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/rancher"

	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

const ProviderNotFound = "provider not found, must be one of ['aws'], got %s"

type SpawnerService interface {
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
	GetWorkspacesCost(context.Context, *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error)

	RegisterWithRancher(context.Context, *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error)
	WriteCredential(context.Context, *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error)
	ReadCredential(context.Context, *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error)
}

//spawnerService manage provider and clusters
type spawnerService struct {
	awsController Controller
	logger        *zap.SugaredLogger
	config        *config.Config

	proto.UnimplementedSpawnerServiceServer
}

//New return ClusterController
func New(logger *zap.SugaredLogger, config *config.Config) SpawnerService {

	svc := &spawnerService{
		awsController: aws.NewAWSController(logger, config),
		logger:        logger,
		config:        config,
	}
	return svc
}

func (s *spawnerService) controller(provider string) (Controller, error) {
	switch provider {
	case "aws":
		return s.awsController, nil
	}
	return nil, fmt.Errorf(ProviderNotFound, provider)
}

//CreateCluster create cluster on the provider specified in request
func (s *spawnerService) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}

	return provider.CreateCluster(ctx, req)
}

//GetCluster get cluster on the providerr specified in request
func (s *spawnerService) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetCluster(ctx, req)
}

//GetClusters get the available clusters in the given provider
func (s *spawnerService) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {

	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetClusters(ctx, req)
}

//AddToken deprecated as of now
func (s *spawnerService) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {

	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.AddToken(ctx, req)
}

//GetToken return the kube token for the cluster in given provider
func (s *spawnerService) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetToken(ctx, req)
}

//AddRoute53Record
func (s *spawnerService) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.AddRoute53Record(ctx, req)
}

//ClusterStatus get cluster status in given provider
func (s *spawnerService) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.ClusterStatus(ctx, req)
}

//AddNode adds new node to the cluster on the provider
func (s *spawnerService) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.AddNode(ctx, req)
}

//DeleteCluster deletes empty cluster on the provider, fails when cluster has nodegroup
func (s *spawnerService) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.DeleteCluster(ctx, req)
}

//DeleteNode deletes node on the given provider cluster
func (s *spawnerService) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.DeleteNode(ctx, req)
}

//CreateVolume create new volume on the provider
func (s *spawnerService) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.CreateVolume(ctx, req)
}

//DeleteVolume delete the volumne on the provider
func (s *spawnerService) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.DeleteVolume(ctx, req)
}

//CreateSnapshot
func (s *spawnerService) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.CreateSnapshot(ctx, req)
}

//CreateSnapshotAndDelete
func (s *spawnerService) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.CreateSnapshotAndDelete(ctx, req)
}

//GetWorkspaceCost returns workspace cost grouped by given group
func (s *spawnerService) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetWorkspacesCost(ctx, req)
}

//RegisterWithRancher register cluster on the rancher, returns the kube manifest to apply on the cluster
func (s *spawnerService) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {

	clusterName := req.ClusterName
	s.logger.Info("registering cluster with rancher ", req.ClusterName)

	client, err := rancher.CreateRancherClient(s.config.RancherAddr, s.config.RancherUsername, s.config.RancherPassword)

	if err != nil {
		s.logger.Error("failed to get rancher client ", client)

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
		s.logger.Errorf("failed to create a rancher cluster '%s' %s", clusterName, err.Error())
		return nil, err
	}

	registrationToken, err := client.ClusterRegistrationToken.Create(&rnchrClient.ClusterRegistrationToken{
		ClusterID: registeredCluster.ID,
	})

	if err != nil {
		//TODO: we may want to revert the creation process,
		//but we will keep it now, so we can manually deal with the registration in case of failure.

		s.logger.Errorf("failed to fetch registration token for '%s' %s", clusterName, err.Error())
		return nil, err
	}
	s.logger.Infof("cluster created on the rancher, apply the manifest file on the target cluster '%s'", registrationToken.ManifestURL)

	return &proto.RancherRegistrationResponse{
		ClusterName: registeredCluster.Name,
		ClusterID:   registrationToken.ClusterID,
		ManifestURL: registrationToken.ManifestURL,
	}, nil

}

//WriteCredential
func (s *spawnerService) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {

	region := s.config.SecretHostRegion
	id := req.GetAccessKeyID()
	key := req.GetSecretAccessKey()
	account := req.GetAccount()

	err := s.writeCredentials(ctx, region, account, id, key)
	if err != nil {
		s.logger.Errorw("failed to save credentials", "error", err, "account", account)
		return nil, err
	}
	return &proto.WriteCredentialResponse{}, nil

}

//ReadCredential
func (s *spawnerService) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {

	region := s.config.SecretHostRegion
	account := req.GetAccount()

	creds, err := s.getCredentials(ctx, region, account)
	if err != nil {
		s.logger.Errorw("failed to get the credentials", "account", account)
		return nil, err
	}
	s.logger.Debugw("credentials found", "account", account, "accessKeyID", creds.AccessKeyID)
	return &proto.ReadCredentialResponse{
		Account:         account,
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
	}, nil
}
