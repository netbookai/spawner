package rancher

import (
	"testing"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var sugar = logger.Sugar()

type SessionManagerClientMock struct {
	GetClusterWithNameMock func(clusterName string) (*rnchrClient.ClusterCollection, error)
	GetAllClustersMock     func() (*rnchrClient.ClusterCollection, error)
	UpdateClusterMock      func(cluster *rnchrClient.Cluster, updateJson map[string]interface{}) (*rnchrClient.Cluster, error)
	GetCloudCredentialMock func(credName string) (*rnchrClient.CloudCredentialCollection, error)
	CreateClusterMock      func(cluster *rnchrClient.Cluster) (*rnchrClient.Cluster, error)
	GetKubeConfigMock      func(cluster *rnchrClient.Cluster) (*rnchrClient.GenerateKubeConfigOutput, error)
	CreateTokenMock        func(newTokenVar *rnchrClient.Token) (*rnchrClient.Token, error)
	DeleteClusterMock      func(cluster *rnchrClient.Cluster) error
}

func (sm SessionManagerClientMock) GetClusterWithName(clusterName string) (*rnchrClient.ClusterCollection, error) {
	return sm.GetClusterWithNameMock(clusterName)
}

func (sm SessionManagerClientMock) GetAllClusters() (*rnchrClient.ClusterCollection, error) {
	return sm.GetAllClustersMock()
}

func (sm SessionManagerClientMock) UpdateCluster(cluster *rnchrClient.Cluster, updateJson map[string]interface{}) (*rnchrClient.Cluster, error) {
	return sm.UpdateClusterMock(cluster, updateJson)
}

func (sm SessionManagerClientMock) GetCloudCredential(credName string) (*rnchrClient.CloudCredentialCollection, error) {
	return sm.GetCloudCredentialMock(credName)
}

func (sm SessionManagerClientMock) CreateCluster(cluster *rnchrClient.Cluster) (*rnchrClient.Cluster, error) {
	return sm.CreateClusterMock(cluster)
}

func (sm SessionManagerClientMock) GetKubeConfig(cluster *rnchrClient.Cluster) (*rnchrClient.GenerateKubeConfigOutput, error) {
	return sm.GetKubeConfigMock(cluster)
}

func (sm SessionManagerClientMock) CreateToken(newTokenVar *rnchrClient.Token) (*rnchrClient.Token, error) {
	return sm.CreateTokenMock(newTokenVar)
}

func (sm SessionManagerClientMock) DeleteCluster(cluster *rnchrClient.Cluster) error {
	return sm.DeleteClusterMock(cluster)
}

func TestGetCluster(t *testing.T) {

	ms := &SessionManagerClientMock{func(clusterName string) (*rnchrClient.ClusterCollection, error) {
		r := rnchrClient.ClusterCollection{}
		testCluster := rnchrClient.Cluster{}
		testCluster.Name = clusterName
		r.Data = append(r.Data, testCluster)
		return &r, nil
	}, nil, nil, nil, nil, nil, nil, nil}
	svc := NewRancherController(ms, nil, sugar)

	clusterName := "test"

	resp, err := svc.GetCluster(clusterName)

	if err != nil {
		t.Errorf("error in calling get cluster: %s", err)
	}
	if resp.Name != clusterName {
		t.Errorf("expected name '%s', got '%s'", clusterName, resp.Name)
	}
}

func TestGetClusterID(t *testing.T) {

	ms := &SessionManagerClientMock{func(clusterName string) (*rnchrClient.ClusterCollection, error) {
		r := rnchrClient.ClusterCollection{}
		testCluster := rnchrClient.Cluster{}
		testCluster.Name = clusterName
		testCluster.ID = "1"
		r.Data = append(r.Data, testCluster)
		return &r, nil
	}, nil, nil, nil, nil, nil, nil, nil}
	svc := NewRancherController(ms, nil, sugar)

	clusterName := "test"
	clusterID := "1"

	resp, err := svc.GetClusterID(clusterName)

	if err != nil {
		t.Errorf("error in calling get cluster id: %s", err)
	}
	if resp != clusterID {
		t.Errorf("expected cluster id '%s', got '%s'", clusterID, resp)
	}
}

func TestGetEksClustersInRegion(t *testing.T) {

	ms := &SessionManagerClientMock{nil, func() (*rnchrClient.ClusterCollection, error) {
		r := rnchrClient.ClusterCollection{}
		testCluster := rnchrClient.Cluster{}
		testCluster.Name = "test"
		testCluster.ID = "1"
		eksConfig := rnchrClient.EKSClusterConfigSpec{}
		eksConfig.Region = "us-west-1"
		testCluster.EKSConfig = &eksConfig
		r.Data = append(r.Data, testCluster)
		testCluster2 := rnchrClient.Cluster{}
		testCluster2.Name = "test2"
		testCluster2.ID = "2"
		eksConfig2 := rnchrClient.EKSClusterConfigSpec{}
		eksConfig2.Region = "us-west-1"
		testCluster2.EKSConfig = &eksConfig2
		r.Data = append(r.Data, testCluster2)
		testCluster3 := rnchrClient.Cluster{}
		testCluster3.Name = "test3"
		testCluster3.ID = "3"
		eksConfig3 := rnchrClient.EKSClusterConfigSpec{}
		eksConfig3.Region = "us-west-2"
		testCluster3.EKSConfig = &eksConfig3
		r.Data = append(r.Data, testCluster3)
		return &r, nil
	}, nil, nil, nil, nil, nil, nil}
	svc := NewRancherController(ms, nil, sugar)

	region := "us-west-1"
	clusterName := "test"
	clusterID := "1"
	clusterName2 := "test2"
	clusterID2 := "2"

	resp, err := svc.GetEksClustersInRegion(region)

	if err != nil {
		t.Errorf("error in calling get eks clusters in region: %s, error : %s", region, err)
	}
	if resp[0].ID != clusterID {
		t.Errorf("expected cluster id '%s', got '%s'", clusterID, resp[0].ID)
	}
	if resp[0].Name != clusterName {
		t.Errorf("expected cluster name '%s', got '%s'", clusterName, resp[0].Name)
	}
	if resp[1].ID != clusterID2 {
		t.Errorf("expected cluster id '%s', got '%s'", clusterID2, resp[1].ID)
	}
	if resp[1].Name != clusterName2 {
		t.Errorf("expected cluster name '%s', got '%s'", clusterName2, resp[1].Name)
	}
}

func TestUpdateCluster(t *testing.T) {

	ms := &SessionManagerClientMock{nil, nil, func(cluster *rnchrClient.Cluster, updateJson map[string]interface{}) (*rnchrClient.Cluster, error) {
		cluster.Name = updateJson["name"].(string)
		cluster.ID = updateJson["id"].(string)
		return cluster, nil
	}, nil, nil, nil, nil, nil}
	svc := NewRancherController(ms, nil, sugar)

	clusterName := "test"
	clusterID := "1"
	updateJson := map[string]interface{}{
		"name": clusterName,
		"id":   clusterID,
	}
	r := rnchrClient.Cluster{}
	newClusterSpec := rnchrClient.ClusterSpec{}

	resp, err := svc.UpdateCluster(&r, newClusterSpec, updateJson)

	if err != nil {
		t.Errorf("error in calling update eks cluster: %s", err)
	}
	if resp.ID != clusterID {
		t.Errorf("expected cluster id '%s', got '%s'", clusterID, resp.ID)
	}
	if resp.Name != clusterName {
		t.Errorf("expected cluster id '%s', got '%s'", clusterName, resp.Name)
	}

}

func TestGetKubeConfig(t *testing.T) {

	ms := &SessionManagerClientMock{func(clusterName string) (*rnchrClient.ClusterCollection, error) {
		r := rnchrClient.ClusterCollection{}
		testCluster := rnchrClient.Cluster{}
		testCluster.Name = clusterName
		r.Data = append(r.Data, testCluster)
		return &r, nil
	}, nil, nil, nil, nil, func(cluster *rnchrClient.Cluster) (*rnchrClient.GenerateKubeConfigOutput, error) {
		kubeconfig := rnchrClient.GenerateKubeConfigOutput{cluster.Name}
		return &kubeconfig, nil
	}, nil, nil}
	svc := NewRancherController(ms, nil, sugar)

	clusterName := "test"

	resp, err := svc.GetKubeConfig(clusterName)

	if err != nil {
		t.Errorf("error in calling get kube config: %s", err)
	}
	if resp != clusterName {
		t.Errorf("expected output '%s', got '%s'", clusterName, resp)
	}
}
