package rancher

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher/eks"
	"gitlab.com/netbook-devs/spawner-service/pkg/util"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

type RancherController struct {
	spawnerServiceRancher RancherClient
	config                *util.Config
	logger                *zap.SugaredLogger
}

func NewRancherController(spawnerServiceRancher RancherClient, config *util.Config, logger *zap.SugaredLogger) RancherController {

	return RancherController{spawnerServiceRancher, config, logger}
}

func (svc RancherController) GetClusterInternal(clusterName string) (*rnchrClient.Cluster, error) {
	clusterSpecList, err := svc.spawnerServiceRancher.GetClusterWithName(clusterName)

	if err != nil || len(clusterSpecList.Data) <= 0 {
		return &rnchrClient.Cluster{}, fmt.Errorf("no cluster found with clustername %s", clusterName)
	}

	clusterSpec := clusterSpecList.Data[0]

	return &clusterSpec, nil
}

func (svc RancherController) GetClusterID(clusterName string) (string, error) {
	clusterSpecList, err := svc.spawnerServiceRancher.GetClusterWithName(clusterName)

	if err != nil || len(clusterSpecList.Data) <= 0 {
		return ("no cluster found with clustername " + clusterName), err
	}

	clusterSpec := clusterSpecList.Data[0].ID

	return clusterSpec, err
}

func (svc RancherController) GetClusterNodes(clusterId string) ([]rnchrClient.Node, error) {
	nodesList, err := svc.spawnerServiceRancher.ListNodes(clusterId)

	if err != nil {
		svc.logger.Errorw("error getting cluster nodes", "clusterid", clusterId, "error", err)
		return []rnchrClient.Node{}, err
	}

	return nodesList.Data, err
}

func (svc RancherController) GetEksClustersInRegion(region string) ([]rnchrClient.Cluster, error) {
	clusterSpecList, err := svc.spawnerServiceRancher.GetAllClusters()
	if err != nil || len(clusterSpecList.Data) <= 0 {
		svc.logger.Errorw("error getting eks clusters in region", "region", region, "error", err)
		return []rnchrClient.Cluster{}, fmt.Errorf("error finding eks clusters")
	}

	clustersInRegion := []rnchrClient.Cluster{}

	for _, clust := range clusterSpecList.Data {
		if clust.EKSConfig != nil && clust.EKSConfig.Region == region {
			clustersInRegion = append(clustersInRegion, clust)
		}
	}

	svc.logger.Info("got eks clusters in region", "region", region, "clusters", clustersInRegion)

	return clustersInRegion, nil
}

func (svc RancherController) UpdateCluster(cluster *rnchrClient.Cluster, clusterConfigPatch rnchrClient.ClusterSpec, appendJson map[string]interface{}) (*rnchrClient.Cluster, error) {

	var finalJson map[string]interface{}

	configJson, err := json.Marshal(clusterConfigPatch)
	if err != nil {
		svc.logger.Errorw("error marshaling clusterconfigpatch", "clusterid", cluster.ID, "error", err)
		return &rnchrClient.Cluster{}, fmt.Errorf("error marshaling clusterconfigpatch")
	}
	json.Unmarshal(configJson, &finalJson)
	if appendJson != nil {
		tempJson, _ := json.Marshal(appendJson)
		json.Unmarshal(tempJson, &finalJson)
	}

	respCluster, err := svc.spawnerServiceRancher.UpdateCluster(cluster, finalJson)
	svc.logger.Infow("in UpdateCluster method", "respCluster", respCluster)

	if err != nil {
		svc.logger.Errorw("error updating cluster", "clusterid", cluster.ID, "error", err)
		return &rnchrClient.Cluster{}, fmt.Errorf("error updating cluster %s", cluster.Name)
	}

	return respCluster, err
}

func (svc RancherController) GetCloudCreds(credName string) (rnchrClient.CloudCredential, error) {
	list, _ := svc.spawnerServiceRancher.GetCloudCredential(credName)

	if len(list.Data) <= 0 {
		svc.logger.Errorw("could not find credential with name", "name", credName)
		return rnchrClient.CloudCredential{}, fmt.Errorf("could not find credential with name %s", credName)
	}

	return list.Data[0], nil
}

func (svc RancherController) AddNodeInternal(nodeSpawnRequest *pb.NodeSpawnRequest) (*rnchrClient.Cluster, error) {
	cluster, err := svc.GetClusterInternal(nodeSpawnRequest.ClusterName)

	if err != nil {
		svc.logger.Errorw("error getting cluster internal", "cluster", nodeSpawnRequest.ClusterName, "error", err)
		return &rnchrClient.Cluster{}, fmt.Errorf("error getting cluster %v", err)
	}

	newClusterSpec, err := eks.AddNodeGroup(cluster, nodeSpawnRequest, map[string]string{})

	if err != nil {
		svc.logger.Errorw("error adding node group to eks cluster", "cluster", nodeSpawnRequest.ClusterName, "error", err)
		return &rnchrClient.Cluster{}, fmt.Errorf("error adding node %v", err)
	}

	cluster, err = svc.UpdateCluster(cluster, newClusterSpec, map[string]interface{}{
		"name": cluster.AppliedSpec.EKSConfig.DisplayName})

	if err != nil {
		svc.logger.Errorw("error updating cluster", "cluster", nodeSpawnRequest.ClusterName, "error", err)
		return &rnchrClient.Cluster{}, fmt.Errorf("error updating cluster %v", err)
	}

	return cluster, err
}

func (svc RancherController) CreateClusterInternal(clusterName string, clusterReq *pb.ClusterRequest) (*rnchrClient.Cluster, error) {
	awsCred, _ := svc.GetCloudCreds(svc.config.AwsCredName)

	eksClustersInRegion, err := svc.GetEksClustersInRegion(clusterReq.Region)

	if err != nil {
		svc.logger.Errorw("error creating cluster failed at getting clusters in region", "cluster", clusterName, "clusterrequest", clusterReq, "error", err)
		return &rnchrClient.Cluster{}, fmt.Errorf("creating cluster failed at getting clusters in region with err %s", err)
	}

	var subnets []string
	if len(eksClustersInRegion) > 0 {
		subnets = *eksClustersInRegion[0].EKSConfig.Subnets

		if len(subnets) <= 0 {
			subnets = eksClustersInRegion[0].EKSStatus.Subnets
		}
	}

	newCluster := eks.CreateCluster(
		awsCred,
		clusterReq,
		clusterName,
		map[string]string{"name": clusterName, "creator": "spawner-service", "provisioner": "rancher"},
		map[string]string{"creator": "spawner-service", "provisioner": "rancher"},
		subnets)

	cluster, err := svc.spawnerServiceRancher.CreateCluster(newCluster)

	if err != nil {
		svc.logger.Errorw("error creating new cluster", "error", err)
		return &rnchrClient.Cluster{}, err
	}
	svc.logger.Infow("new cluster created", "cluster", clusterName)

	return cluster, nil
}

func (svc RancherController) GetClusterStatusInternal(req *pb.ClusterStatusRequest) (string, error) {
	cluster, err := svc.GetClusterInternal(req.ClusterName)

	if err != nil {
		svc.logger.Errorw("error getting cluster status internal", "cluster", req.ClusterName, "clusterstatusrequest", req, "error", err)
		return "", err
	}

	return cluster.State, nil
}

func (svc RancherController) GetKubeConfig(clusterName string) (string, error) {
	cluster, err := svc.GetClusterInternal(clusterName)

	if err != nil {
		return "", fmt.Errorf("error getting cluster with name %s", clusterName)
	}

	yaml, err := svc.spawnerServiceRancher.GetKubeConfig(cluster)

	if err != nil {
		return "", fmt.Errorf("error getting kube config for cluster with name %s", clusterName)
	}

	return yaml.Config, nil
}

func (svc RancherController) CreateToken(clusterName string, region string) (string, error) {
	clusterId, err := svc.GetClusterID(clusterName)

	if err != nil {
		svc.logger.Errorw("error getting clusterid", "cluster", clusterName, "error", err)
		return "", err
	}

	existingTokens, listErr := svc.spawnerServiceRancher.ListTokens(clusterId)

	if listErr != nil {
		svc.logger.Warnw("error getting tokens for cluster", "cluster", clusterName, "clusterId", clusterId, "error", listErr)
	}

	for _, tok := range existingTokens.Data {
		if tok.ClusterID == clusterId {
			svc.logger.Warnw("token already exists for cluster", "cluster", clusterName, "clusterId", clusterId, "region", region)
			return "token already exists for cluster", nil
		}
	}

	newTokenVar := &rnchrClient.Token{
		ClusterID:   clusterId,
		TTLMillis:   2592000000,
		Description: "Automated Token for " + clusterName,
	}
	newToken, err := svc.spawnerServiceRancher.CreateToken(newTokenVar)

	if err != nil {
		svc.logger.Errorw("error creating token ", "clustername", clusterName)
		return "", err
	}

	status, err := aws.CreateAwsSecret(clusterName, clusterId, newToken.Token, region)
	if err != nil {
		svc.logger.Errorw("error creating new aws secret", "cluster", clusterName, "clusterid", clusterId, "region", region, "error", err)
	}

	return status, err
}

func (svc RancherController) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	var clusterName string
	if clusterName = req.ClusterName; len(clusterName) == 0 {
		clusterName = fmt.Sprintf("%s-%s", req.Provider, req.Region)
	}

	cluster, err := svc.CreateClusterInternal(clusterName, req)

	if err != nil {
		svc.logger.Errorw("error creating cluster ", "clustername", clusterName)
		return &pb.ClusterResponse{
			Error: err.Error(),
		}, status.Errorf(codes.Internal, "error creating cluster")
	}

	_, err = svc.CreateToken(clusterName, req.Region)
	if err != nil {
		svc.logger.Errorw("error creating cluster token", "clustername", clusterName)
		_, delErr := svc.DeleteCluster(ctx, &pb.ClusterDeleteRequest{
			ClusterName: clusterName,
		})
		if delErr != nil {
			svc.logger.Errorw("error deleting cluster", "cluster", clusterName, "error", delErr)
		}

		return &pb.ClusterResponse{
			Error: err.Error(),
		}, status.Errorf(codes.Internal, "error creating cluster token")
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
		svc.logger.Errorw("error getting AWS secret", "cluster", req.ClusterName, "region", req.Region, "error", err)
		return &pb.GetTokenResponse{
			Token: "",
			Error: "error getting AWSSecret",
		}, status.Errorf(codes.Internal, "error getting Aws secret")
	}

	return &pb.GetTokenResponse{
		Token:         token,
		RancherServer: svc.config.RancherAddr,
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
	cluster, err := svc.GetClusterInternal(req.ClusterName)

	if err != nil {
		return &pb.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	err = svc.spawnerServiceRancher.DeleteCluster(cluster)

	if err != nil {
		return &pb.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.ClusterDeleteResponse{}, nil
}

func (svc RancherController) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	cluster, err := svc.GetClusterInternal(req.ClusterName)

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

func (svc RancherController) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {
	cluster, err := svc.GetClusterInternal(req.ClusterName)

	svc.logger.Infow("got cluster in getcluster", "cluster", req.ClusterName, "clusterobj", cluster)

	nodes, err := svc.GetClusterNodes(cluster.ID)
	if err != nil {
		svc.logger.Errorw("error getting nodes for cluster", "cluster", req.ClusterName, "clusterobj", cluster, "error", err)
		return &pb.ClusterSpec{}, fmt.Errorf("error getting nodes for clustername %s", req.ClusterName)
	}

	var nodeSpecList []*pb.NodeSpec
	for _, node := range nodes {
		nodeSpecList = append(nodeSpecList, &pb.NodeSpec{
			Name: node.Name,
			// TODO: Sid add disksize
			Instance: node.Labels["node.kubernetes.io/instance-type"],
			// DiskSize: ,
			HostName:         node.Hostname,
			State:            node.State,
			Uuid:             node.UUID,
			IpAddr:           node.IPAddress,
			ClusterId:        node.ClusterID,
			Labels:           node.Labels,
			Availabilityzone: node.Labels["topology.kubernetes.io/zone"],
		})
	}

	resp := pb.ClusterSpec{
		Name:      cluster.Name,
		ClusterId: cluster.ID,
		NodeSpec:  nodeSpecList,
	}

	return &resp, nil
}

func (svc RancherController) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {
	if req.Provider == "aws" && req.Scope == "public" {
		clusters, err := svc.GetEksClustersInRegion(req.Region)

		if err != nil {
			svc.logger.Errorw("error getting cluster in getclusters", "getclustersrequest", req, "error", err)
			return &pb.GetClustersResponse{}, err
		}

		resp := pb.GetClustersResponse{
			Clusters: [](*pb.ClusterSpec){},
		}
		for _, cluster := range clusters {
			nodes := []*pb.NodeSpec{}

			for _, node := range *cluster.EKSConfig.NodeGroups {
				nodes = append(nodes, &pb.NodeSpec{
					Name:     *node.NodegroupName,
					Instance: *node.InstanceType,
					DiskSize: int32(*node.DiskSize),
				})
			}

			resp.Clusters = append(resp.Clusters, &pb.ClusterSpec{
				Name:     cluster.Name,
				NodeSpec: nodes,
			})
		}

		return &resp, nil
	} else {
		svc.logger.Errorw("provider or scope not supported yet", "getclustersrequest", req)
		return &pb.GetClustersResponse{}, fmt.Errorf("provider %s not supported yet", req.Provider)
	}
}
