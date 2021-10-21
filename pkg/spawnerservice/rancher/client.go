package rancher

import (
	rnchrTypes "github.com/rancher/norman/types"
	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

type SpawnerServiceRancher struct {
	rancherClient *rnchrClient.Client
}

type RancherClient interface {
	GetClusterWithName(clusterName string) (*rnchrClient.ClusterCollection, error)
	GetAllClusters() (*rnchrClient.ClusterCollection, error)
	UpdateCluster(cluster *rnchrClient.Cluster, updateJson map[string]interface{}) (*rnchrClient.Cluster, error)
	GetCloudCredential(credName string) (*rnchrClient.CloudCredentialCollection, error)
	CreateCluster(cluster *rnchrClient.Cluster) (*rnchrClient.Cluster, error)
	GetKubeConfig(cluster *rnchrClient.Cluster) (*rnchrClient.GenerateKubeConfigOutput, error)
	CreateToken(newTokenVar *rnchrClient.Token) (*rnchrClient.Token, error)
	DeleteCluster(cluster *rnchrClient.Cluster) error
}

func NewSpawnerServiceClient(rancherClient *rnchrClient.Client) RancherClient {

	return SpawnerServiceRancher{rancherClient}
}

func (svc SpawnerServiceRancher) GetClusterWithName(clusterName string) (*rnchrClient.ClusterCollection, error) {
	clusterSpecList, err := svc.rancherClient.Cluster.ListAll(
		&rnchrTypes.ListOpts{
			Filters: map[string]interface{}{"name": clusterName},
		},
	)
	return clusterSpecList, err
}

func (svc SpawnerServiceRancher) GetAllClusters() (*rnchrClient.ClusterCollection, error) {
	clusterSpecList, err := svc.rancherClient.Cluster.ListAll(
		&rnchrTypes.ListOpts{},
	)
	return clusterSpecList, err
}

func (svc SpawnerServiceRancher) UpdateCluster(cluster *rnchrClient.Cluster, updateJson map[string]interface{}) (*rnchrClient.Cluster, error) {
	respCluster, err := svc.rancherClient.Cluster.Update(cluster, updateJson)
	return respCluster, err
}

func (svc SpawnerServiceRancher) GetCloudCredential(credName string) (*rnchrClient.CloudCredentialCollection, error) {
	list, err := svc.rancherClient.CloudCredential.ListAll(
		&rnchrTypes.ListOpts{
			Filters: map[string]interface{}{"name": credName},
		},
	)
	return list, err
}

func (svc SpawnerServiceRancher) CreateCluster(cluster *rnchrClient.Cluster) (*rnchrClient.Cluster, error) {

	return svc.rancherClient.Cluster.Create(cluster)
}

func (svc SpawnerServiceRancher) GetKubeConfig(cluster *rnchrClient.Cluster) (*rnchrClient.GenerateKubeConfigOutput, error) {

	yaml, err := svc.rancherClient.Cluster.ActionGenerateKubeconfig(cluster)

	return yaml, err
}

func (svc SpawnerServiceRancher) CreateToken(newTokenVar *rnchrClient.Token) (*rnchrClient.Token, error) {
	newToken, err := svc.rancherClient.Token.Create(newTokenVar)

	return newToken, err
}

func (svc SpawnerServiceRancher) DeleteCluster(cluster *rnchrClient.Cluster) error {
	err := svc.rancherClient.Cluster.Delete(cluster)

	return err
}
