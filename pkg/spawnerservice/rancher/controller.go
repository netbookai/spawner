package rancher

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher/eks"
	"gitlab.com/netbook-devs/spawner-service/pkg/util"

	rnchrTypes "github.com/rancher/norman/types"
	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

type RancherController struct {
	rancherClient *rnchrClient.Client
	config        *util.Config
	logger        *zap.SugaredLogger
}

func NewRancherController(logger *zap.SugaredLogger, config util.Config) RancherController {

	rancherClient, _ := common.CreateRancherClient(config.RancherAddr, config.RancherUsername, config.RancherPassword)

	return RancherController{rancherClient, &config, logger}
}

func (svc RancherController) GetCluster(clusterName string) (*rnchrClient.Cluster, error) {
	clusterSpecList, err := svc.rancherClient.Cluster.ListAll(
		&rnchrTypes.ListOpts{
			Filters: map[string]interface{}{"name": clusterName},
		},
	)

	if err != nil || len(clusterSpecList.Data) <= 0 {
		return &rnchrClient.Cluster{}, fmt.Errorf("no cluster found with clustername %s", clusterName)
	}

	clusterSpec := clusterSpecList.Data[0]

	return &clusterSpec, nil
}

func (svc RancherController) GetClusterID(clusterName string) (string, error) {
	clusterSpecList, err := svc.rancherClient.Cluster.ListAll(
		&rnchrTypes.ListOpts{
			Filters: map[string]interface{}{"name": clusterName},
		},
	)

	if err != nil || len(clusterSpecList.Data) <= 0 {
		return ("no cluster found with clustername " + clusterName), err
	}

	clusterSpec := clusterSpecList.Data[0].ID

	return clusterSpec, err
}

func (svc RancherController) GetEksClustersInRegion(region string) ([]rnchrClient.Cluster, error) {
	clusterSpecList, err := svc.rancherClient.Cluster.ListAll(
		&rnchrTypes.ListOpts{
			// Filters: map[string]interface{}{"provider": "eks"},
		},
	)

	if err != nil || len(clusterSpecList.Data) <= 0 {
		return []rnchrClient.Cluster{}, fmt.Errorf("error finding eks clusters")
	}

	clustersInRegion := []rnchrClient.Cluster{}

	for _, clust := range clusterSpecList.Data {
		if clust.EKSConfig != nil && clust.EKSConfig.Region == region {
			clustersInRegion = append(clustersInRegion, clust)
		}
	}

	return clustersInRegion, nil
}

func (svc RancherController) UpdateCluster(cluster *rnchrClient.Cluster, clusterConfigPatch rnchrClient.ClusterSpec, appendJson map[string]interface{}) (*rnchrClient.Cluster, error) {

	var finalJson map[string]interface{}

	configJson, err := json.Marshal(clusterConfigPatch)
	if err != nil {
		return &rnchrClient.Cluster{}, fmt.Errorf("error marshaling clusterconfigpatch")
	}
	json.Unmarshal(configJson, &finalJson)
	if appendJson != nil {
		tempJson, _ := json.Marshal(appendJson)
		json.Unmarshal(tempJson, &finalJson)
	}

	respCluster, err := svc.rancherClient.Cluster.Update(cluster, finalJson)

	svc.logger.Infow("respCluster", respCluster)

	if err != nil {
		return &rnchrClient.Cluster{}, fmt.Errorf("error updating cluster %s", cluster.Name)
	}

	return respCluster, err
}

func (svc RancherController) GetCloudCreds(credName string) (rnchrClient.CloudCredential, error) {
	list, _ := svc.rancherClient.CloudCredential.ListAll(
		&rnchrTypes.ListOpts{
			Filters: map[string]interface{}{"name": credName},
		},
	)

	if len(list.Data) <= 0 {
		return rnchrClient.CloudCredential{}, fmt.Errorf("could not find credential with name %s", credName)
	}

	return list.Data[0], nil
}

func (svc RancherController) AddNodeInternal(nodeSpawnRequest *pb.NodeSpawnRequest) (*rnchrClient.Cluster, error) {
	cluster, err := svc.GetCluster(nodeSpawnRequest.ClusterName)

	if err != nil {
		return &rnchrClient.Cluster{}, fmt.Errorf("error getting cluster %v", err)
	}

	// nodeGroupCount := len(*cluster.EKSConfig.NodeGroups)
	// newNodeGroupName := "ng-" + strconv.Itoa(nodeGroupCount+1)
	newClusterSpec, err := eks.AddNodeGroup(cluster, nodeSpawnRequest, map[string]string{})

	if err != nil {
		return &rnchrClient.Cluster{}, fmt.Errorf("error adding node %v", err)
	}

	cluster, err = svc.UpdateCluster(cluster, newClusterSpec, map[string]interface{}{
		"name": cluster.AppliedSpec.EKSConfig.DisplayName})

	if err != nil {
		return &rnchrClient.Cluster{}, fmt.Errorf("error updating cluster %v", err)
	}

	return cluster, err
}

func (svc RancherController) CreateClusterInternal(clusterName string, clusterReq *pb.ClusterRequest) (*rnchrClient.Cluster, error) {
	awsCred, _ := svc.GetCloudCreds(svc.config.AwsCredName)

	eksClustersInRegion, err := svc.GetEksClustersInRegion(clusterReq.Region)

	if err != nil {
		return &rnchrClient.Cluster{}, fmt.Errorf("creating cluster failed at getting clusters in region with err %s", err)
	}

	var subnets []string
	if len(eksClustersInRegion) > 0 {
		subnets = *eksClustersInRegion[0].EKSConfig.Subnets

		if len(subnets) <= 0 {
			subnets = eksClustersInRegion[0].EKSStatus.Subnets
		}
	}

	newCluster := eks.CreateCluster(awsCred, clusterReq, clusterName+"-eks-"+strconv.Itoa(len(eksClustersInRegion)+1), map[string]string{}, map[string]string{}, subnets)

	cluster, err := svc.rancherClient.Cluster.Create(newCluster)

	if err != nil {
		svc.logger.Errorw("error", err)
		return &rnchrClient.Cluster{}, err
	}

	return cluster, nil
}

func (svc RancherController) GetClusterStatusInternal(req *pb.ClusterStatusRequest) (string, error) {
	cluster, err := svc.GetCluster(req.ClusterName)

	if err != nil {
		return "", err
	}

	return cluster.State, nil
}

func (svc RancherController) GetKubeConfig(clusterName string) (string, error) {
	cluster, err := svc.GetCluster(clusterName)

	if err != nil {
		return "", fmt.Errorf("error getting cluster with name %s", clusterName)
	}

	yaml, err := svc.rancherClient.Cluster.ActionGenerateKubeconfig(cluster)

	if err != nil {
		return "", fmt.Errorf("error getting kube config for cluster with name %s", clusterName)
	}

	return yaml.Config, nil
}

func (svc RancherController) CreateToken(clusterName string, region string) (string, error) {
	clusterID, err := svc.GetClusterID(clusterName)

	newTokenVar := &rnchrClient.Token{
		ClusterID:   clusterID,
		TTLMillis:   2592000000,
		Description: "Automated Token for " + clusterName,
	}
	newToken, err := svc.rancherClient.Token.Create(newTokenVar)

	status, err := aws.CreateAwsSecret(clusterName, clusterID, newToken.Token, region)

	return status, err
}

func (svc RancherController) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	clusterName := fmt.Sprintf("%s-%s", req.Provider, req.Region)

	cluster, err := svc.CreateClusterInternal(clusterName, req)

	if err != nil {
		svc.logger.Errorw("error creating cluster ", "clustername", clusterName)
		return &pb.ClusterResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.ClusterResponse{
		ClusterName:   cluster.Name,
		NodeGroupName: *(*cluster.EKSConfig.NodeGroups)[0].NodegroupName,
		Error:         "",
	}, nil
}

func (svc RancherController) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	status, err := svc.CreateToken(req.ClusterName, req.Region)

	if err != nil {
		return &pb.AddTokenResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.AddTokenResponse{
		Status: status,
	}, err
}

func (svc RancherController) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	token, err := aws.GetAwsSecret(req.ClusterName, req.Region)

	if err != nil {
		return &pb.GetTokenResponse{
			Token: "",
		}, err
	}

	return &pb.GetTokenResponse{
		Token: token,
	}, err
}

func (svc RancherController) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	status, err := svc.GetClusterStatusInternal(req)

	if err != nil {
		return &pb.ClusterStatusResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.ClusterStatusResponse{
		Status: status,
	}, nil
}

func (svc RancherController) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	_, err := svc.AddNodeInternal(req)

	if err != nil {
		return &pb.NodeSpawnResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.NodeSpawnResponse{}, nil
}

func (svc RancherController) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	cluster, err := svc.GetCluster(req.ClusterName)

	if err != nil {
		return &pb.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	err = svc.rancherClient.Cluster.Delete(cluster)

	if err != nil {
		return &pb.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.ClusterDeleteResponse{}, nil
}

func (svc RancherController) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	cluster, err := svc.GetCluster(req.ClusterName)

	if err != nil {
		return &pb.NodeDeleteResponse{
			Error: err.Error(),
		}, err
	}

	newClusterSpec, err := eks.DeleteNodeGroup(cluster, req)
	if err != nil {
		return &pb.NodeDeleteResponse{
			Error: err.Error(),
		}, err
	}

	_, err = svc.UpdateCluster(cluster, newClusterSpec, map[string]interface{}{
		"name": cluster.AppliedSpec.EKSConfig.DisplayName})

	if err != nil {
		return &pb.NodeDeleteResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.NodeDeleteResponse{}, nil
}

func (svc RancherController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	return &pb.CreateVolumeResponse{}, nil
}

func (svc RancherController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	return &pb.DeleteVolumeResponse{}, nil
}

func (svc RancherController) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	return &pb.CreateSnapshotResponse{}, nil
}

func (svc RancherController) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	return &pb.CreateSnapshotAndDeleteResponse{}, nil
}
