package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWS_CLUSTER_ROLE_NAME    = "netbook-AWS-ServiceRoleForEKS-BADBEEF"
	AWS_NODE_GROUP_ROLE_NAME = "netbook-AWS-NodeGroupInstanceRole-CAFE"
	//cluster role policy
	EKS_CLUSTER_POLICY_ARN = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
	EKS_SERVICE_POLICY_ARN = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"

	//node group role policy
	EKS_WORKER_NODE_POLICY_ARN      = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
	EKS_EC2_CONTAINER_RO_POLICY_ARN = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
	EKS_CNI_POLICY_ARN              = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"

	EKS_ASSUME_ROLE_DOC = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":["eks.amazonaws.com"]},"Action":["sts:AssumeRole"]}]}`
	EC2_ASSUME_ROLE_DOC = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":["ec2.amazonaws.com"]},"Action":["sts:AssumeRole"]}]}`
)

var (
	ERR_NODEGROUP_EXIST = errors.New("nodegroup already exist")
	ERR_NO_NODEGROUP    = errors.New("no nodegroup exist in cluster")
)

type AWSController struct {
	logger *zap.SugaredLogger
}

//NewAWSController
func NewAWSController(logger *zap.SugaredLogger) *AWSController {
	return &AWSController{
		logger: logger,
	}
}

//getClusterSpec Get the cluster spec of given name
func getClusterSpec(ctx context.Context, client *eks.EKS, name string) (*eks.Cluster, error) {
	input := eks.DescribeClusterInput{
		Name: &name,
	}
	resp, err := client.DescribeClusterWithContext(ctx, &input)
	return resp.Cluster, err
}

//CreateCluster Create new cluster with given specification, no op if cluste already exist
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

	ctrl.logger.Debugf("checking cluster status for '%s', region '%s'", clusterName, region)

	cluster, err := getClusterSpec(ctx, eksClient, clusterName)

	if err != nil {
		if err.(awserr.Error).Code() == eks.ErrCodeResourceNotFoundException {

			ctrl.logger.Debugf("cluster '%s' does not exist, creating ...", clusterName)
			cluster, err = ctrl.createClusterInternal(ctx, session, clusterName, req)
			if err != nil {
				ctrl.logger.Error("failed to create clsuter '%s' %s", clusterName, err.Error())
				return nil, err
			}

			ctrl.logger.Info("cluster '%s' is creating state, it might take some time, please check AWS console for status", clusterName)
		}
	} else {
		ctrl.logger.Infof("cluster '%s', already exist", clusterName)
	}

	return &proto.ClusterResponse{
		ClusterName: *cluster.Name,
	}, nil
}

func (ctrl AWSController) getNodeHealth(ctx context.Context, client *eks.EKS, cluster, nodeName string) (*eks.NodegroupHealth, error) {

	node, err := client.DescribeNodegroupWithContext(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   &cluster,
		NodegroupName: &nodeName,
	})

	if err != nil {
		return nil, err
	}
	return node.Nodegroup.Health, nil
}

func healthProto(health *eks.NodegroupHealth) *proto.Health {
	pr := &proto.Health{}
	issues := make([]*proto.Issue, 0, len(health.Issues))

	for _, is := range health.Issues {
		rids := make([]string, 0, 5)
		for _, r := range is.ResourceIds {
			rids = append(rids, *r)
		}

		issues = append(issues, &proto.Issue{
			Code:        *is.Code,
			Description: *is.Message,
			ResourceIds: rids,
		})
	}
	pr.Issue = issues
	return pr
}

//GetCluster Describe cluster with the given name and region
func (ctrl AWSController) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	response := &proto.ClusterSpec{}
	region := req.Region
	clusterName := req.ClusterName
	accountName := req.AccountName
	session, err := NewSession(ctx, region, accountName)

	ctrl.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		ctrl.logger.Error("failed to fetch cluster status", err)
		return nil, err
	}

	k8sClient, err := session.getK8sClient(cluster)
	if err != nil {
		ctrl.logger.Error(" Failed to create kube client ", err)
		return nil, err
	}
	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	response.Name = clusterName

	if err != nil {
		ctrl.logger.Error(" Failed to query node list ", err)
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
			ctrl.logger.Errorw("failed to get the health check", "error", err)
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
		ctrl.logger.Error("failed to list clusters", err)
		return &proto.GetClustersResponse{}, err
	}

	resp := proto.GetClustersResponse{
		Clusters: [](*proto.ClusterSpec){},
	}

	for _, cluster := range listClusterOut.Clusters {

		clusterSpec, err := getClusterSpec(ctx, client, *cluster)

		if err != nil {
			ctrl.logger.Errorw("failed to get cluster details", "cluster", *cluster, "error", err)
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
		if ScopeTag() != *scope {
			//skip clusters which is of not spawner env scope
			continue
		}

		input := &eks.ListNodegroupsInput{ClusterName: cluster}
		nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
		if err != nil {
			ctrl.logger.Errorf("failed to fetch nodegroups %s", err.Error())
		}

		nodes := []*proto.NodeSpec{}
		for _, cNodeGroup := range nodeGroupList.Nodegroups {
			input := &eks.DescribeNodegroupInput{
				NodegroupName: cNodeGroup,
				ClusterName:   cluster}
			nodeGroupDetails, err := client.DescribeNodegroupWithContext(ctx, input)

			if err != nil {
				ctrl.logger.Error("failed to fetch nodegroups details ", *cNodeGroup)
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

func (ctrl AWSController) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	region := req.Region
	clusterName := req.ClusterName
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	ctrl.logger.Debugw("fetching cluster status", "cluster-name", clusterName, "region", region)
	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		ctrl.logger.Errorw("failed to fetch cluster status", "error", err, "cluster", clusterName, "region", region)
		return &proto.ClusterStatusResponse{
			Error: err.Error(),
		}, err
	}

	return &proto.ClusterStatusResponse{
		Status: *cluster.Status,
	}, err
}

//getDefaultNode Get any existing node from the cluster as default node
//if node with `newNode` exist return error
func (ctrl AWSController) getDefaultNode(ctx context.Context, client *eks.EKS, clusterName, nodeName string) (*eks.Nodegroup, error) {

	input := &eks.ListNodegroupsInput{ClusterName: &clusterName}
	nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
	if err != nil {
		ctrl.logger.Errorf("failed to fetch nodegroups: %s", err.Error())
		return nil, err
	}

	if len(nodeGroupList.Nodegroups) == 0 {
		return nil, ERR_NO_NODEGROUP
	}

	for _, nodeGroup := range nodeGroupList.Nodegroups {
		if *nodeGroup == nodeName {
			return nil, ERR_NODEGROUP_EXIST
		}
	}

	nodeDetailsinput := &eks.DescribeNodegroupInput{
		NodegroupName: nodeGroupList.Nodegroups[0],
		ClusterName:   &clusterName}
	nodeGroupDetails, err := client.DescribeNodegroupWithContext(ctx, nodeDetailsinput)

	return nodeGroupDetails.Nodegroup, err

}

func (ctrl AWSController) getNewNodeGroupSpecFromCluster(ctx context.Context, session *Session, cluster *eks.Cluster, nodeSpec *proto.NodeSpec) (*eks.CreateNodegroupInput, error) {

	iamClient := session.getIAMClient()

	roleName := AWS_NODE_GROUP_ROLE_NAME
	nodeRole, newRole, err := ctrl.createRoleOrGetExisting(ctx, iamClient, roleName, "node group instance policy role", EC2_ASSUME_ROLE_DOC)

	if err != nil {
		ctrl.logger.Errorf("failed to create node group role '%s' %w", AWS_NODE_GROUP_ROLE_NAME, err)
		return nil, err
	}

	if newRole {

		err = ctrl.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_WORKER_NODE_POLICY_ARN)

		if err != nil {
			ctrl.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_WORKER_NODE_POLICY_ARN, AWS_NODE_GROUP_ROLE_NAME, err)
			return nil, err
		}

		err = ctrl.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_EC2_CONTAINER_RO_POLICY_ARN)

		if err != nil {
			ctrl.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_EC2_CONTAINER_RO_POLICY_ARN, AWS_NODE_GROUP_ROLE_NAME, err)
			return nil, err
		}

		err = ctrl.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_CNI_POLICY_ARN)

		if err != nil {
			ctrl.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_CNI_POLICY_ARN, AWS_NODE_GROUP_ROLE_NAME, err)
			return nil, err
		}
	}

	diskSize := int64(nodeSpec.DiskSize)

	labels := getNodeLabel(nodeSpec)

	amiType := ""
	//Choose Amazon Linux 2 (AL2_x86_64) for Linux non-GPU instances, Amazon Linux 2 GPU Enabled (AL2_x86_64_GPU) for Linux GPU instances
	if nodeSpec.GpuEnabled {
		ctrl.logger.Infof("requested gpu node for '%s'", nodeSpec.Name)
		amiType = "AL2_x86_64_GPU"
	} else {
		amiType = "AL2_x86_64"
	}

	return &eks.CreateNodegroupInput{
		AmiType:       &amiType,
		CapacityType:  common.StrPtr("ON_DEMAND"),
		NodeRole:      nodeRole.Arn,
		InstanceTypes: []*string{&nodeSpec.Instance},
		ClusterName:   cluster.Name,
		DiskSize:      &diskSize,
		NodegroupName: &nodeSpec.Name,
		Labels:        labels,
		Subnets:       cluster.ResourcesVpcConfig.SubnetIds,
		ScalingConfig: &eks.NodegroupScalingConfig{
			DesiredSize: common.Int64Ptr(1),
			MinSize:     common.Int64Ptr(1),
			MaxSize:     common.Int64Ptr(1),
		},
		Tags: labels,
	}, nil

}

func (ctrl AWSController) getNodeSpecFromDefault(defaultNode *eks.Nodegroup, clusterName string, nodeSpec *proto.NodeSpec) *eks.CreateNodegroupInput {
	diskSize := int64(nodeSpec.DiskSize)

	//add labels from the given spec
	labels := getNodeLabel(nodeSpec)

	amiType := ""
	//Choose Amazon Linux 2 (AL2_x86_64) for Linux non-GPU instances, Amazon Linux 2 GPU Enabled (AL2_x86_64_GPU) for Linux GPU instances
	if nodeSpec.GpuEnabled {
		ctrl.logger.Infof("requested gpu node for '%s'", nodeSpec.Name)
		amiType = "AL2_x86_64_GPU"
	} else {
		amiType = "AL2_x86_64"
	}

	return &eks.CreateNodegroupInput{
		AmiType:        &amiType,
		CapacityType:   defaultNode.CapacityType,
		NodeRole:       defaultNode.NodeRole,
		InstanceTypes:  []*string{&nodeSpec.Instance},
		ClusterName:    &clusterName,
		DiskSize:       &diskSize,
		NodegroupName:  &nodeSpec.Name,
		ReleaseVersion: defaultNode.ReleaseVersion,
		Labels:         labels,
		Subnets:        defaultNode.Subnets,
		ScalingConfig: &eks.NodegroupScalingConfig{
			DesiredSize: common.Int64Ptr(1),
			MinSize:     common.Int64Ptr(1),
			MaxSize:     common.Int64Ptr(1),
		},
		Tags: labels,
	}
}

//AddNode adds new node group to the existing cluster, cluster atleast have 1 node group already present
func (ctrl AWSController) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {

	//create a new node on the given cluster with the NodeSpec
	clusterName := req.ClusterName
	region := req.Region
	nodeSpec := req.NodeSpec

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {

		ctrl.logger.Errorw("unable to get cluster, spec", "error", err.Error(), "cluster", clusterName, "region", region)
		return nil, err
	}

	ctrl.logger.Infof("querying default nodes on cluster '%s' in region '%s'", clusterName, region)
	defaultNode, err := ctrl.getDefaultNode(ctx, client, clusterName, nodeSpec.Name)

	var newNodeGroupInput *eks.CreateNodegroupInput

	if err != nil {
		if errors.Is(err, ERR_NODEGROUP_EXIST) {
			return nil, err
		}

		if errors.Is(err, ERR_NO_NODEGROUP) {
			//no node group present,
			ctrl.logger.Infof("default nodegroup not found in cluster '%s', creating NodegroupRequest from cluster config ", clusterName)
			newNodeGroupInput, err = ctrl.getNewNodeGroupSpecFromCluster(ctx, session, cluster, nodeSpec)
			if err != nil {
				return nil, err
			}
		}
	} else {
		ctrl.logger.Infof("found default nodegroup '%s' in cluster '%s', creating NodegroupRequest from default node config", *defaultNode.NodegroupName, clusterName)
		newNodeGroupInput = ctrl.getNodeSpecFromDefault(defaultNode, clusterName, nodeSpec)
	}

	out, err := client.CreateNodegroupWithContext(ctx, newNodeGroupInput)
	if err != nil {
		ctrl.logger.Errorf("failed to add a node '%s': %w", nodeSpec.Name, err)
		return nil, err
	}
	ctrl.logger.Infof("creating nodegroup '%s' on cluster '%s', Status : %s, it might take some time. Please check AWS console.", nodeSpec.Name, clusterName, *out.Nodegroup.Status)
	return &proto.NodeSpawnResponse{}, err
}

func (ctrl AWSController) deleteAllNodegroups(ctx context.Context, client *eks.EKS, clusterName string) error {
	input := &eks.ListNodegroupsInput{ClusterName: &clusterName}
	nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
	if err != nil {
		return err
	}

	if len(nodeGroupList.Nodegroups) == 0 {
		return nil
	}

	for _, nodeGroupName := range nodeGroupList.Nodegroups {
		//drop em
		if err = ctrl.deleteNode(ctx, client, clusterName, *nodeGroupName); err != nil {
			return err
		}
	}
	ctrl.logger.Infow("all attached nodegroups are being deleted")
	return nil
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

	//get node groups attached to clients when force delete is enabled.
	//if available delete all attached node groups and proceed to deleting cluster
	if forceDelete {
		err = ctrl.deleteAllNodegroups(ctx, client, clusterName)
		if err != nil {
			ctrl.logger.Errorw("failed to delete attached nodegroups", "error", err)
			return nil, err
		}

	}

	//FIXME : cannot delete the cluster untill all nodes are dropped.
	deleteOut, err := client.DeleteClusterWithContext(ctx, &eks.DeleteClusterInput{
		Name: &clusterName,
	})

	if err != nil {
		ctrl.logger.Errorf("failed to delete cluster '%s': %s", clusterName, err.Error())
		return &proto.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	ctrl.logger.Infof("requested cluster '%s' to be deleted, Status :%s. It might take some time, check AWS console for more.", clusterName, *deleteOut.Cluster.Status)

	return &proto.ClusterDeleteResponse{}, nil
}

func (ctrl AWSController) deleteNode(ctx context.Context, client *eks.EKS, cluster, node string) error {
	nodeDeleteOut, err := client.DeleteNodegroupWithContext(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   &cluster,
		NodegroupName: &node,
	})

	if err != nil {
		return err

	}
	ctrl.logger.Infof("requested nodegroup '%s' to be deleted, Status %s. It might take some time, check AWS console for more.", node, *nodeDeleteOut.Nodegroup.Status)
	return nil
}

//DeleteNode
func (ctrl AWSController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	clusterName := req.ClusterName
	nodeName := req.NodeGroupName
	region := req.Region

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	err = ctrl.deleteNode(ctx, client, clusterName, nodeName)
	if err != nil {
		ctrl.logger.Errorw("failed to delete nodegroup", "nodename", nodeName)
		return &proto.NodeDeleteResponse{Error: err.Error()}, err
	}

	return &proto.NodeDeleteResponse{}, nil
}

func (ctrl AWSController) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	return &proto.AddTokenResponse{}, nil
}

func (ctrl AWSController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {

	region := req.Region
	clusterName := req.ClusterName

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()
	ctrl.logger.Debugw("fetching cluster status", "cluster", clusterName, "region", region)

	cluster, err := getClusterSpec(ctx, client, clusterName)
	if err != nil {
		ctrl.logger.Errorw("failed to get cluster spec", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	kubeConfig, err := session.getKubeConfig(cluster)
	if err != nil {
		ctrl.logger.Errorw("failed to get k8s config", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}
	return &proto.GetTokenResponse{
		Token:    kubeConfig.BearerToken,
		CaData:   string(kubeConfig.CAData),
		Endpoint: kubeConfig.Host,
	}, nil
}
