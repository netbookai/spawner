package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//getClusterSpec Get the cluster spec of given name
func getClusterSpec(ctx context.Context, client *eks.EKS, name string) (*eks.Cluster, error) {
	input := eks.DescribeClusterInput{
		Name: &name,
	}
	resp, err := client.DescribeClusterWithContext(ctx, &input)
	return resp.Cluster, err
}

func (svc AWSController) createClusterInternal(ctx context.Context, session *Session, clusterName string, req *proto.ClusterRequest) (*eks.Cluster, error) {

	var subnetIds []*string
	region := session.Region

	awsRegionNetworkStack, err := GetRegionWkspNetworkStack(session)
	if err != nil {
		svc.logger.Error(ctx, "error getting network stack for region", "region", region, "error", err)
		return nil, err
	}

	if awsRegionNetworkStack.Vpc != nil && len(awsRegionNetworkStack.Subnets) > 0 {
		for _, subn := range awsRegionNetworkStack.Subnets {
			subnetIds = append(subnetIds, subn.SubnetId)
		}
		svc.logger.Info(ctx, "got network stack for region", "vpc", awsRegionNetworkStack.Vpc.VpcId, "subnets", subnetIds)
	} else {
		awsRegionNetworkStack, err = CreateRegionWkspNetworkStack(session)
		if err != nil {
			svc.logger.Error(ctx, "error creating network stack for region with no clusters", "region", region, "error", err)
			svc.logger.Warn(ctx, "rolling back network stack changes as creation failed", "region", region)
			delErr := DeleteRegionWkspNetworkStack(session, *awsRegionNetworkStack)
			if delErr != nil {
				svc.logger.Error(ctx, "error deleting network stack for region", "region", region, "error", delErr)
			}

			return nil, err
		}
		for _, subn := range awsRegionNetworkStack.Subnets {
			subnetIds = append(subnetIds, subn.SubnetId)
		}
		svc.logger.Info(ctx, "created network stack for region", "vpc", awsRegionNetworkStack.Vpc.VpcId, "subnets", subnetIds)
	}

	tags := labels.DefaultTags()
	tags[constants.ClusterNameLabel] = &clusterName
	//override with additional labels from request
	for k, v := range req.Labels {
		v := v
		tags[k] = &v
	}

	iamClient := session.getIAMClient()
	roleName := AWS_CLUSTER_ROLE_NAME

	eksRole, newRole, err := svc.createRoleOrGetExisting(ctx, iamClient, roleName, "eks cluster and service access role", EKS_ASSUME_ROLE_DOC)

	if err != nil {
		svc.logger.Error(ctx, "failed to create role %w", err)
		return nil, err
	}

	if newRole {
		err = svc.attachPolicy(ctx, iamClient, roleName, EKS_CLUSTER_POLICY_ARN)
		if err != nil {
			svc.logger.Error(ctx, "failed to attach policy '%s' to role '%s' %w", EKS_CLUSTER_POLICY_ARN, roleName, err)
			return nil, err
		}

		err = svc.attachPolicy(ctx, iamClient, roleName, EKS_SERVICE_POLICY_ARN)
		if err != nil {
			svc.logger.Error(ctx, "failed to attach policy '%s' to role '%s' %w", EKS_SERVICE_POLICY_ARN, roleName, err)
			return nil, err
		}
	}

	clusterInput := &eks.CreateClusterInput{
		Name: &clusterName,
		ResourcesVpcConfig: &eks.VpcConfigRequest{
			SubnetIds:             subnetIds,
			EndpointPublicAccess:  aws.Bool(true),
			EndpointPrivateAccess: aws.Bool(false),
		},
		Tags: tags,
		//		Version: &constants.KubeVersion,
		RoleArn: eksRole.Arn,
	}

	client := session.getEksClient()
	createClusterOutput, err := client.CreateClusterWithContext(ctx, clusterInput)
	if err != nil {
		svc.logger.Error(ctx, "failed to create cluster", "error", err)
		return nil, err
	}

	return createClusterOutput.Cluster, nil

}

//isExist check if the given cluster exist in EKS
//
// This function currently uses non paginated version of ListClusters, safe to assume that we would not have clusters more than 100 in single account.
func isExist(ctx context.Context, client *eks.EKS, name string) (bool, error) {

	res, err := client.ListClustersWithContext(ctx, &eks.ListClustersInput{})
	if err != nil {
		return false, err
	}

	for _, cluster := range res.Clusters {
		if name == *cluster {
			return true, nil
		}
	}
	return false, nil
}

//CreateCluster Create new cluster with given specification, no op if cluster already exist
func (ctrl AWSController) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {

	var clusterName string
	if clusterName = req.ClusterName; len(clusterName) == 0 {
		clusterName = fmt.Sprintf("%s-%s", req.Provider, req.Region)
	}

	region := req.Region
	accountName := req.AccountName
	session, err := NewSession(ctx, region, accountName)

	if err != nil {
		return nil, err
	}
	eksClient := session.getEksClient()

	ctrl.logger.Debug(ctx, "checking cluster status for '%s', region '%s'", clusterName, region)
	exist, err := isExist(ctx, eksClient, clusterName)

	if err != nil {
		return nil, errors.Wrap(err, "CreateCluster failed")
	}

	if exist {
		ctrl.logger.Info(ctx, "cluster '%s', already exist", clusterName)
		return nil, errors.New("cluster already exist")
	}

	ctrl.logger.Debug(ctx, "cluster '%s' does not exist, creating ...", clusterName)
	cluster, err := ctrl.createClusterInternal(ctx, session, clusterName, req)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to create cluster ", "cluster", clusterName, "error", err)
		return nil, err
	}

	ctrl.logger.Info(ctx, "cluster is in creating state, it might take some time, please check AWS console for status", "cluster", clusterName)

	return &proto.ClusterResponse{
		ClusterName: *cluster.Name,
	}, nil
}

//GetClusters return active, spawner created clusters
func (ctrl AWSController) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {

	//get all clusters in given region
	region := req.Region
	accountName := req.AccountName
	session, err := NewSession(ctx, region, accountName)
	if err != nil {
		return nil, err
	}

	client := session.getEksClient()

	//list cluster allows paginated query,
	listClusterInput := &eks.ListClustersInput{}
	listClusterOut, err := client.ListClustersWithContext(ctx, listClusterInput)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to list clusters", err)
		return &proto.GetClustersResponse{}, err
	}

	resp := proto.GetClustersResponse{
		Clusters: [](*proto.ClusterSpec){},
	}

	for _, cluster := range listClusterOut.Clusters {

		clusterSpec, err := getClusterSpec(ctx, client, *cluster)

		if err != nil {
			ctrl.logger.Error(ctx, "failed to get cluster details", "cluster", *cluster, "error", err)
			continue

		}
		creator, ok := clusterSpec.Tags[constants.CreatorLabel]
		if !ok {
			//unknown creator
			continue
		}

		if *clusterSpec.Status != "ACTIVE" || *creator != constants.SpawnerServiceLabel {
			continue
		}

		scope, ok := clusterSpec.Tags[constants.Scope]
		if !ok {
			continue
		}
		if labels.ScopeTag() != *scope {
			//skip clusters which is of not spawner env scope
			continue
		}

		input := &eks.ListNodegroupsInput{ClusterName: cluster}
		nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
		if err != nil {
			ctrl.logger.Error(ctx, "failed to fetch nodegroups %s", err.Error())
		}

		nodes := []*proto.NodeSpec{}
		for _, cNodeGroup := range nodeGroupList.Nodegroups {
			input := &eks.DescribeNodegroupInput{
				NodegroupName: cNodeGroup,
				ClusterName:   cluster}
			nodeGroupDetails, err := client.DescribeNodegroupWithContext(ctx, input)

			if err != nil {
				ctrl.logger.Error(ctx, "failed to fetch nodegroups details ", *cNodeGroup)
				continue
			}

			node := &proto.NodeSpec{Name: *cNodeGroup}

			if nodeGroupDetails.Nodegroup.InstanceTypes != nil {
				node.Instance = *nodeGroupDetails.Nodegroup.InstanceTypes[0]
			}
			if nodeGroupDetails.Nodegroup.DiskSize != nil {
				node.DiskSize = int32(*nodeGroupDetails.Nodegroup.DiskSize)
			}

			node.Health = healthProto(nodeGroupDetails.Nodegroup.Health)
			nodes = append(nodes, node)
		}

		resp.Clusters = append(resp.Clusters, &proto.ClusterSpec{
			Name:     *cluster,
			NodeSpec: nodes,
		})
	}
	return &resp, nil
}

//GetCluster Describe cluster with the given name and region
func (ctrl AWSController) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	response := &proto.ClusterSpec{}
	region := req.Region
	clusterName := req.ClusterName
	accountName := req.AccountName
	session, err := NewSession(ctx, region, accountName)

	ctrl.logger.Debug(ctx, "fetching cluster status for '%s', region '%s'", clusterName, region)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		ctrl.logger.Error(ctx, "failed to fetch cluster status", err)
		return nil, err
	}

	k8sClient, err := session.getK8sClient(cluster)
	if err != nil {
		ctrl.logger.Error(ctx, " Failed to create kube client ", err)
		return nil, err
	}
	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	response.Name = clusterName

	if err != nil {
		ctrl.logger.Error(ctx, "failed to query node list ", err)
		return nil, err
	}

	var nodeSpecList []*proto.NodeSpec
	for _, node := range nodeList.Items {
		nodeGroupName := node.Labels[constants.NodeNameLabel]
		addresses := node.Status.Addresses
		ipAddr := ""
		hostName := node.Name
		for _, address := range addresses {
			switch address.Type {

			case "InternalIP":
				ipAddr = address.Address
			case "HostName":
				hostName = address.Address
			}
		}

		state := "inactive"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" {
				state = "active"
			}
		}

		ephemeralStorage := node.Status.Capacity.StorageEphemeral()

		//get node health
		var nodeHealth *proto.Health
		health, err := ctrl.getNodeHealth(ctx, client, clusterName, nodeGroupName)
		if err != nil {
			ctrl.logger.Error(ctx, "failed to get the health check", "error", err)
		} else {
			nodeHealth = healthProto(health)
		}

		//we will use MB for the disk size, int32 is too small for bytes
		diskSize := ephemeralStorage.Value() / 1024 / 1024
		nodeSpecList = append(nodeSpecList, &proto.NodeSpec{
			Name: nodeGroupName,
			//ClusterId:        node.ClusterID,
			Instance:         node.Labels["node.kubernetes.io/instance-type"],
			DiskSize:         int32(diskSize),
			HostName:         hostName,
			State:            state,
			Uuid:             string(node.ObjectMeta.UID),
			IpAddr:           ipAddr,
			Labels:           node.Labels,
			Availabilityzone: node.Labels["topology.kubernetes.io/zone"],
			Health:           nodeHealth,
		})
	}
	response.NodeSpec = nodeSpecList
	return response, nil
}

//ClusterStatus get the cluster status
func (ctrl AWSController) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	region := req.Region
	clusterName := req.ClusterName
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	ctrl.logger.Debug(ctx, "fetching cluster status", "cluster-name", clusterName, "region", region)
	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		ctrl.logger.Error(ctx, "failed to fetch cluster status", "error", err, "cluster", clusterName, "region", region)
		return &proto.ClusterStatusResponse{
			Error: err.Error(),
		}, err
	}

	return &proto.ClusterStatusResponse{
		Status: *cluster.Status,
	}, err
}

//DeleteCluster delete empty cluster, cluster should not have any nodegroup attached.
func (ctrl AWSController) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {

	clusterName := req.ClusterName
	region := req.Region
	forceDelete := req.ForceDelete

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		err := errors.New(err.(awserr.Error).Message())
		return nil, errors.Wrap(err, "DeleteCluster: ")
	}

	if scope, ok := cluster.Tags[constants.Scope]; !ok || *scope != labels.ScopeTag() {
		return nil, fmt.Errorf("cluster doesnt not available in '%s'", labels.ScopeTag())
	}

	//get node groups attached to clients when force delete is enabled.
	//if available delete all attached node groups and proceed to deleting cluster
	if forceDelete {
		ctrl.logger.Info(ctx, "force deleting all nodegroups of cluster", "cluster", clusterName)
		err = ctrl.deleteAllNodegroups(ctx, client, clusterName)
		if err != nil {
			ctrl.logger.Error(ctx, "failed to delete attached nodegroups", "error", err)
			return nil, err
		}

		ctrl.logger.Info(ctx, "waiting for all nodegroups deletion", "cluster", clusterName)
		err = ctrl.waitForAllNodegroupsDeletion(ctx, client, clusterName)
		if err != nil {
			ctrl.logger.Error(ctx, "failed waiting for deletion of attached nodegroups", "error", err)
			return nil, err
		}
		ctrl.logger.Info(ctx, "done waiting for all nodegroups to delete", "cluster", clusterName)
	}

	deleteOut, err := client.DeleteClusterWithContext(ctx, &eks.DeleteClusterInput{
		Name: &clusterName,
	})

	if err != nil {
		ctrl.logger.Error(ctx, "failed to delete cluster '%s': %s", clusterName, err.Error())
		return &proto.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	ctrl.logger.Info(ctx, "requested cluster to be deleted. It might take some time, check AWS console for more.", "cluster", clusterName, "status", *deleteOut.Cluster.Status)

	return &proto.ClusterDeleteResponse{}, nil
}
