package service

import (
	"context"
	"fmt"

	"github.com/netbookai/log"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	aws "gitlab.com/netbook-devs/spawner-service/pkg/service/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/azure"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/gcp"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/rancher"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	"gitlab.com/netbook-devs/spawner-service/pkg/types"

	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

const ProviderNotFound = "provider not found, must be one of ['aws', 'azure'], got %s"

type SpawnerService interface {
	CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error)
	GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error)
	GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error)
	AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error)
	GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error)
	ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error)
	AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error)
	DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error)
	DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error)
	CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error)
	CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error)
	DeleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error)
	CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error)
	GetWorkspacesCost(context.Context, *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error)
	GetApplicationsCost(context.Context, *proto.GetApplicationsCostRequest) (*proto.GetApplicationsCostResponse, error)
	GetKubeConfig(ctx context.Context, in *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error)
	TagNodeInstance(ctx context.Context, req *proto.TagNodeInstanceRequest) (*proto.TagNodeInstanceResponse, error)

	RegisterWithRancher(context.Context, *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error)
	WriteCredential(context.Context, *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error)
	ReadCredential(context.Context, *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error)
	AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error)
	GetCostByTime(ctx context.Context, req *proto.GetCostByTimeRequest) (*proto.GetCostByTimeResponse, error)

	GetContainerRegistryAuth(ctx context.Context, in *proto.GetContainerRegistryAuthRequest) (*proto.GetContainerRegistryAuthResponse, error)
	CreateContainerRegistryRepo(ctx context.Context, in *proto.CreateContainerRegistryRepoRequest) (*proto.CreateContainerRegistryRepoResponse, error)

	RegisterClusterOIDC(ctx context.Context, in *proto.RegisterClusterOIDCRequest) (*proto.RegisterClusterOIDCResponse, error)
	CreateRoute53Records(ctx context.Context, req *proto.CreateRoute53RecordsRequest) (*proto.CreateRoute53RecordsResponse, error)
	GetRoute53TXTRecords(ctx context.Context, req *proto.GetRoute53TXTRecordsRequest) (*proto.GetRoute53TXTRecordsResponse, error)
	DeleteRoute53Records(ctx context.Context, req *proto.DeleteRoute53RecordsRequest) (*proto.DeleteRoute53RecordsResponse, error)
}

//spawnerService manage provider and clusters
type spawnerService struct {
	awsController   Controller
	azureController Controller
	gcpController   Controller
	logger          log.Logger
	proto.UnimplementedSpawnerServiceServer
}

//New return ClusterController
func New(logger log.Logger) SpawnerService {

	svc := &spawnerService{
		awsController:   aws.NewAWSController(logger),
		azureController: azure.NewController(logger),
		gcpController:   gcp.NewController(logger),
		logger:          logger,
	}
	return svc
}

func (s *spawnerService) controller(provider string) (Controller, error) {
	switch provider {
	case "aws":
		return s.awsController, nil
	case "azure":
		return s.azureController, nil
	case "gcp":
		return s.gcpController, nil
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

//GetApplicationsCost returns workspace cost grouped by given group
func (s *spawnerService) GetApplicationsCost(ctx context.Context, req *proto.GetApplicationsCostRequest) (*proto.GetApplicationsCostResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetApplicationsCost(ctx, req)
}

//RegisterWithRancher register cluster on the rancher, returns the kube manifest to apply on the cluster
func (s *spawnerService) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {

	clusterName := req.ClusterName
	s.logger.Info(ctx, "registering cluster with rancher ", "cluster", req.ClusterName)

	conf := config.Get()
	client, err := rancher.CreateRancherClient(conf.RancherAddr, conf.RancherUsername, conf.RancherPassword)

	if err != nil {
		s.logger.Error(ctx, "failed to get rancher client ", client)

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
		s.logger.Error(ctx, "failed to create a rancher cluster", "cluster", clusterName, "error", err.Error())
		return nil, err
	}

	registrationToken, err := client.ClusterRegistrationToken.Create(&rnchrClient.ClusterRegistrationToken{
		ClusterID: registeredCluster.ID,
	})

	if err != nil {
		//TODO: we may want to revert the creation process,
		//but we will keep it now, so we can manually deal with the registration in case of failure.

		s.logger.Error(ctx, "failed to fetch registration token ", "cluster", clusterName, "error", err.Error())
		return nil, err
	}
	s.logger.Info(ctx, "cluster created on the rancher, apply the manifest file on the target cluster", "manifest-url", registrationToken.ManifestURL)

	return &proto.RancherRegistrationResponse{
		ClusterName: registeredCluster.Name,
		ClusterID:   registrationToken.ClusterID,
		ManifestURL: registrationToken.ManifestURL,
	}, nil

}

func validCredType(ct string) bool {
	switch ct {
	case constants.CredAws, constants.CredAzure, constants.CredGitPat:
		return true
	}
	return false
}

//WriteCredential
func (s *spawnerService) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {

	account := req.GetAccount()
	region := config.Get().SecretHostRegion

	credType := req.GetType()

	if !validCredType(credType) {
		return nil, constants.ErrInvalidCredentiualType
	}

	var cred system.Credentials
	cred_type := "unknown"

	switch credType {

	case constants.CredAws:

		cred_type = "AwsCredential"
		if c := req.GetAwsCred(); c != nil {
			cred = &system.AwsCredential{
				Name:   account,
				Id:     c.GetAccessKeyID(),
				Secret: c.GetSecretAccessKey(),
				Token:  c.GetToken(),
			}
		}

	case constants.CredAzure:
		cred_type = "AzureCredential"
		if c := req.GetAzureCred(); c != nil {
			cred = &system.AzureCredential{
				SubscriptionID: c.GetSubscriptionID(),
				TenantID:       c.GetTenantID(),
				ClientID:       c.GetClientID(),
				ClientSecret:   c.GetClientSecret(),
				ResourceGroup:  c.GetResourceGroup(),
				Name:           account,
			}
		}
	case constants.CredGitPat:
		cred_type = "GithubPersonalAccessToken"
		if c := req.GetGitPat(); c != nil {
			cred = &system.GithubPersonalAccessToken{
				Name:  account,
				Token: c.Token,
			}
		}
	case constants.GcpLabel:
		cred_type = "GcpCredential"
		if c := req.GetGcpCred(); c != nil {

			cred = &system.GCPCredential{
				Name:        account,
				ProjectId:   c.GetProjectID(),
				Certificate: c.GetCertificate(),
			}
		}
	default:
		return nil, fmt.Errorf("invalid provider '%s'", credType)
	}

	if cred == nil {
		return nil, fmt.Errorf(" %s credentials must be set for type '%s'", cred_type, credType)

	}

	err := s.writeCredentials(ctx, region, account, credType, cred)
	if err != nil {
		s.logger.Error(ctx, "failed to save credentials", "error", err, "account", account)
		return nil, err
	}
	return &proto.WriteCredentialResponse{}, nil

}

//ReadCredential
func (s *spawnerService) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {

	region := config.Get().SecretHostRegion
	account := req.GetAccount()
	credType := req.GetType()

	if !validCredType(credType) {
		return nil, constants.ErrInvalidCredentiualType
	}

	creds, err := s.getCredentials(ctx, region, account, credType)
	if err != nil {
		s.logger.Error(ctx, "failed to get the credentials", "account", account, "error", err)
		return nil, err
	}
	p := &proto.ReadCredentialResponse{
		Account: account,
	}

	switch credType {
	case constants.CredAws:
		c := creds.GetAws()
		p.Cred = &proto.ReadCredentialResponse_AwsCred{
			AwsCred: &proto.AwsCredentials{
				AccessKeyID:     c.Id,
				SecretAccessKey: c.Secret,
				Token:           c.Token,
			},
		}

	case constants.CredAzure:
		c := creds.GetAzure()
		p.Cred = &proto.ReadCredentialResponse_AzureCred{
			AzureCred: &proto.AzureCredentials{
				SubscriptionID: c.SubscriptionID,
				TenantID:       c.TenantID,
				ClientID:       c.ClientID,
				ClientSecret:   c.ClientSecret,
				ResourceGroup:  c.ResourceGroup,
			},
		}

	case constants.CredGitPat:
		c := creds.GetGitPAT()
		p.Cred = &proto.ReadCredentialResponse_GitPat{
			GitPat: &proto.GithubPersonalAccessToken{
				Token: c.Token,
			},
		}
	case constants.GcpLabel:
		c := creds.GetGcp()
		p.Cred = &proto.ReadCredentialResponse_GcpCred{
			GcpCred: &proto.GcpCredentials{
				ProjectID:   c.ProjectId,
				Certificate: c.Certificate,
			},
		}
	}

	s.logger.Debug(ctx, "credentials found", "account", account, "credential_type", credType)
	return p, nil
}

//AddRoute53Record
func (s *spawnerService) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	dnsName := req.GetDnsName()
	recordName := req.GetRecordName()
	regionName := req.GetRegion()

	isAwsResource := req.Provider == string(constants.AwsCloud)

	changeId, err := s.addRoute53Record(ctx, dnsName, recordName, regionName, isAwsResource)
	if err != nil {
		s.logger.Error(ctx, "failed to add route53 record", "error", err)
		return nil, err
	}
	s.logger.Info(ctx, "added route 53 record", "change-id", changeId)
	return &proto.AddRoute53RecordResponse{}, nil
}

func (s *spawnerService) GetKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetKubeConfig(ctx, req)
}

func (s *spawnerService) TagNodeInstance(ctx context.Context, req *proto.TagNodeInstanceRequest) (*proto.TagNodeInstanceResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.TagNodeInstance(ctx, req)
}

//GetWorkspaceCost returns filtered cost grouped by given group and time
func (s *spawnerService) GetCostByTime(ctx context.Context, req *proto.GetCostByTimeRequest) (*proto.GetCostByTimeResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetCostByTime(ctx, req)
}

func (s *spawnerService) GetContainerRegistryAuth(ctx context.Context, req *proto.GetContainerRegistryAuthRequest) (*proto.GetContainerRegistryAuthResponse, error) {

	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetContainerRegistryAuth(ctx, req)
}

func (s *spawnerService) CreateContainerRegistryRepo(ctx context.Context, req *proto.CreateContainerRegistryRepoRequest) (*proto.CreateContainerRegistryRepoResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.CreateContainerRegistryRepo(ctx, req)
}

func (s *spawnerService) DeleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {
	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.DeleteSnapshot(ctx, req)
}

func (s *spawnerService) RegisterClusterOIDC(ctx context.Context, req *proto.RegisterClusterOIDCRequest) (*proto.RegisterClusterOIDCResponse, error) {

	provider, err := s.controller(req.Provider)
	if err != nil {
		return nil, err
	}
	return provider.RegisterClusterOIDC(ctx, req)
}

func (s *spawnerService) GetRoute53TXTRecords(ctx context.Context, req *proto.GetRoute53TXTRecordsRequest) (*proto.GetRoute53TXTRecordsResponse, error) {
	records, err := s.getRoute53TXTRecords(ctx)
	if err != nil {
		return nil, err
	}

	response := &proto.GetRoute53TXTRecordsResponse{
		Records: make([]*proto.Route53ResourceRecordSet, 0, len(records)),
	}

	for _, r := range records {
		responseRecord := &proto.Route53ResourceRecordSet{
			Type:            r.Type,
			Name:            r.Name,
			ResourceRecords: make([]*proto.Route53ResourceRecord, 0, len(r.ResourceRecords)),
			TtlInSeconds:    r.TTLInSeconds,
		}

		for _, resourceRecord := range r.ResourceRecords {
			responseRecord.ResourceRecords = append(responseRecord.ResourceRecords, &proto.Route53ResourceRecord{
				Value: resourceRecord.Value,
			})
		}

		response.Records = append(response.Records, responseRecord)

	}

	return response, nil

}

func (s *spawnerService) CreateRoute53Records(ctx context.Context, req *proto.CreateRoute53RecordsRequest) (*proto.CreateRoute53RecordsResponse, error) {

	recordsToAdd := make([]types.Route53ResourceRecordSet, 0, len(req.Records))

	for _, r := range req.Records {
		recordToAdd := types.Route53ResourceRecordSet{
			Type:            r.Type,
			Name:            r.Name,
			ResourceRecords: make([]types.ResourceRecordValue, 0, len(r.ResourceRecords)),
			TTLInSeconds:    r.TtlInSeconds,
		}

		for _, resourceRecord := range r.ResourceRecords {
			recordToAdd.ResourceRecords = append(recordToAdd.ResourceRecords, types.ResourceRecordValue{
				Value: resourceRecord.Value,
			})
		}

		recordsToAdd = append(recordsToAdd, recordToAdd)
	}

	err := s.createRoute53Records(ctx, recordsToAdd)
	if err != nil {
		return nil, err
	}

	response := &proto.CreateRoute53RecordsResponse{}

	return response, nil

}

func (s *spawnerService) DeleteRoute53Records(ctx context.Context, req *proto.DeleteRoute53RecordsRequest) (*proto.DeleteRoute53RecordsResponse, error) {

	recordsToDelete := make([]types.Route53ResourceRecordSet, 0, len(req.Records))

	for _, r := range req.Records {
		recordToDelete := types.Route53ResourceRecordSet{
			Type:            r.Type,
			Name:            r.Name,
			ResourceRecords: make([]types.ResourceRecordValue, 0, len(r.ResourceRecords)),
			TTLInSeconds:    r.TtlInSeconds,
		}

		for _, resourceRecord := range r.ResourceRecords {
			recordToDelete.ResourceRecords = append(recordToDelete.ResourceRecords, types.ResourceRecordValue{
				Value: resourceRecord.Value,
			})
		}

		recordsToDelete = append(recordsToDelete, recordToDelete)
	}

	err := s.deleteRoute53Records(ctx, recordsToDelete)
	if err != nil {
		return nil, err
	}

	response := &proto.DeleteRoute53RecordsResponse{}

	return response, nil

}
