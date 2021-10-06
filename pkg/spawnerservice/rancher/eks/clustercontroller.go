package eks

import (
	"fmt"

	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher/common"
)

func AddNodeGroup(cluster *rnchrClient.Cluster, nodeSpawnRequest *pb.NodeSpawnRequest, tags map[string]string) (rnchrClient.ClusterSpec, error) {

	for _, nodeGroup := range *cluster.EKSConfig.NodeGroups {
		if nodeSpawnRequest.NodeSpec.Name == *nodeGroup.NodegroupName {
			return rnchrClient.ClusterSpec{}, fmt.Errorf("nodegroup already exists with name %s", nodeSpawnRequest.NodeSpec.Name)
		}
	}

	newNodeGroup := rnchrClient.NodeGroup{
		DiskSize:             common.Int64Ptr(int64(nodeSpawnRequest.NodeSpec.DiskSize)),
		InstanceType:         common.StrPtr(nodeSpawnRequest.NodeSpec.Instance),
		NodegroupName:        common.StrPtr(nodeSpawnRequest.NodeSpec.Name),
		MinSize:              common.Int64Ptr(1),
		DesiredSize:          common.Int64Ptr(1),
		MaxSize:              common.Int64Ptr(1),
		Gpu:                  common.BoolPtr(false),
		Labels:               &map[string]string{},
		RequestSpotInstances: common.BoolPtr(false),
		ResourceTags:         &map[string]string{},
		Version:              common.StrPtr("1.20"),
		Tags:                 &tags,
		UserData:             common.StrPtr(""),
		Subnets:              (*cluster.EKSConfig.NodeGroups)[0].Subnets,
		Ec2SshKey:            common.StrPtr(""),
	}

	newNodesGroupsList := append((*cluster.EKSConfig.NodeGroups), newNodeGroup)

	newClusterSpec := rnchrClient.ClusterSpec{}
	newClusterSpec.EKSConfig = &rnchrClient.EKSClusterConfigSpec{
		NodeGroups: &newNodesGroupsList,
	}

	return newClusterSpec, nil
}

func DeleteNodeGroup(cluster *rnchrClient.Cluster, nodeDeleteRequest *pb.NodeDeleteRequest) (rnchrClient.ClusterSpec, error) {

	index := -1
	for i, nodeGroup := range *cluster.EKSConfig.NodeGroups {
		if nodeDeleteRequest.NodeGroupName == *nodeGroup.NodegroupName {
			index = i
			break
		}
	}

	if index == -1 {
		return rnchrClient.ClusterSpec{}, fmt.Errorf("no nodegroup with name %s in cluster %s", nodeDeleteRequest.NodeGroupName, nodeDeleteRequest.ClusterName)
	}

	leftNg := (*cluster.EKSConfig.NodeGroups)[:index]
	rightNg := (*cluster.EKSConfig.NodeGroups)[index+1:]

	newNodeGroup := append(leftNg, rightNg...)

	newClusterSpec := rnchrClient.ClusterSpec{}
	newClusterSpec.EKSConfig = &rnchrClient.EKSClusterConfigSpec{
		NodeGroups: &newNodeGroup,
	}

	return newClusterSpec, nil
}

func CreateCluster(awsCred rnchrClient.CloudCredential, clusterReq *pb.ClusterRequest, clusterName string, clusterTags map[string]string, nodeGroupTags map[string]string, subnets []string) *rnchrClient.Cluster {
	newNodeGroup := rnchrClient.NodeGroup{
		DiskSize:             common.Int64Ptr(int64(clusterReq.Node.DiskSize)),
		InstanceType:         common.StrPtr(clusterReq.Node.Instance),
		NodegroupName:        common.StrPtr(clusterReq.Node.Name),
		MinSize:              common.Int64Ptr(1),
		DesiredSize:          common.Int64Ptr(1),
		MaxSize:              common.Int64Ptr(1),
		Gpu:                  common.BoolPtr(false),
		Labels:               &map[string]string{"node": "clusterReq.Node.Name", "creator": "spawner-service"},
		RequestSpotInstances: common.BoolPtr(false),
		ResourceTags:         &map[string]string{},
		Version:              common.StrPtr("1.20"),
		Tags:                 &nodeGroupTags,
		UserData:             common.StrPtr(""),
		Subnets:              &[]string{},
		Ec2SshKey:            common.StrPtr(""),
	}

	newCluster := rnchrClient.Cluster{
		DockerRootDir: *common.StrPtr("/var/lib/docker"),
		Name:          clusterName,
		EKSConfig: &rnchrClient.EKSClusterConfigSpec{
			AmazonCredentialSecret: awsCred.ID,
			DisplayName:            clusterName,
			Imported:               false,
			KmsKey:                 common.StrPtr(""),
			KubernetesVersion:      common.StrPtr("1.20"),
			LoggingTypes:           &[]string{},
			NodeGroups:             &[]rnchrClient.NodeGroup{newNodeGroup},
			PrivateAccess:          common.BoolPtr(false),
			PublicAccess:           common.BoolPtr(true),
			PublicAccessSources:    &[]string{},
			Region:                 clusterReq.Region,
			SecretsEncryption:      common.BoolPtr(false),
			SecurityGroups:         &[]string{},
			ServiceRole:            common.StrPtr(""),
			Subnets:                &subnets,
			Tags:                   &clusterTags,
		},
		WindowsPreferedCluster:  false,
		EnableClusterAlerting:   false,
		EnableClusterMonitoring: false,
		EnableNetworkPolicy:     common.BoolPtr(false),
	}

	return &newCluster
}
