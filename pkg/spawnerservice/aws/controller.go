package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/common"
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
	config *config.Config
}

//NewAWSController
func NewAWSController(logger *zap.SugaredLogger, config *config.Config) *AWSController {
	return &AWSController{
		logger: logger,
		config: config,
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
func (ctrl AWSController) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {

	var clusterName string
	if clusterName = req.ClusterName; len(clusterName) == 0 {
		clusterName = fmt.Sprintf("%s-%s", req.Provider, req.Region)
	}

	region := req.Region
	accountName := req.AccountName
	session, err := NewSession(ctrl.config, region, accountName)

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

	return &pb.ClusterResponse{
		ClusterName: *cluster.Name,
	}, nil
}

//GetCluster Describe cluster with the given name and region
func (ctrl AWSController) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {

	response := &pb.ClusterSpec{}
	region := req.Region
	clusterName := req.ClusterName
	accountName := req.AccountName
	session, err := NewSession(ctrl.config, region, accountName)

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

	var nodeSpecList []*pb.NodeSpec
	for _, node := range nodeList.Items {
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

		//we will use MB for the disk size, int32 is too small for bytes
		diskSize := ephemeralStorage.Value() / 1024 / 1024
		nodeSpecList = append(nodeSpecList, &pb.NodeSpec{
			Name: node.Name,
			//ClusterId:        node.ClusterID,
			Instance:         node.Labels["node.kubernetes.io/instance-type"],
			DiskSize:         int32(diskSize),
			HostName:         hostName,
			State:            state,
			Uuid:             string(node.ObjectMeta.UID),
			IpAddr:           ipAddr,
			Labels:           node.Labels,
			Availabilityzone: node.Labels["topology.kubernetes.io/zone"],
		})
	}

	response.NodeSpec = nodeSpecList

	return response, nil
}

func (ctrl AWSController) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {

	//get all clusters in given region
	region := req.Region
	accountName := req.AccountName
	session, err := NewSession(ctrl.config, region, accountName)
	if err != nil {
		return nil, err
	}

	client := session.getEksClient()

	//list cluster allows paginated query,
	listClusterInput := &eks.ListClustersInput{}
	listClusterOut, err := client.ListClustersWithContext(ctx, listClusterInput)
	if err != nil {
		ctrl.logger.Error("failed to list clusters", err)
		return &pb.GetClustersResponse{}, err
	}

	resp := pb.GetClustersResponse{
		Clusters: [](*pb.ClusterSpec){},
	}

	for _, cluster := range listClusterOut.Clusters {

		//clusterDetails, _ := getClusterSpec(ctx, client, *cluster)
		input := &eks.ListNodegroupsInput{ClusterName: cluster}
		nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
		if err != nil {
			ctrl.logger.Errorf("failed to fetch nodegroups %s", err.Error())
		}

		nodes := []*pb.NodeSpec{}
		for _, cNodeGroup := range nodeGroupList.Nodegroups {
			input := &eks.DescribeNodegroupInput{
				NodegroupName: cNodeGroup,
				ClusterName:   cluster}
			nodeGroupDetails, err := client.DescribeNodegroupWithContext(ctx, input)

			if err != nil {
				ctrl.logger.Error("failed to fetch nodegroups details ", *cNodeGroup)
			}

			node := &pb.NodeSpec{Name: *cNodeGroup}

			if nodeGroupDetails.Nodegroup.InstanceTypes != nil {
				node.Instance = *nodeGroupDetails.Nodegroup.InstanceTypes[0]
			}
			if nodeGroupDetails.Nodegroup.DiskSize != nil {
				node.DiskSize = int32(*nodeGroupDetails.Nodegroup.DiskSize)
			}
			nodes = append(nodes, node)
		}

		resp.Clusters = append(resp.Clusters, &pb.ClusterSpec{
			Name:     *cluster,
			NodeSpec: nodes,
		})
	}

	return &resp, nil
}

func (ctrl AWSController) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	region := req.Region
	clusterName := req.ClusterName
	session, err := NewSession(ctrl.config, region, req.AccountName)

	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	ctrl.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		ctrl.logger.Error("failed to fetch cluster status", err)
		return &pb.ClusterStatusResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.ClusterStatusResponse{
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

func (ctrl AWSController) getNewNodeGroupSpecFromCluster(ctx context.Context, session *Session, cluster *eks.Cluster, nodeSpec *pb.NodeSpec) (*eks.CreateNodegroupInput, error) {

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
	}, nil

}

func (ctrl AWSController) getNodeSpecFromDefault(defaultNode *eks.Nodegroup, clusterName string, nodeSpec *pb.NodeSpec) *eks.CreateNodegroupInput {
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
	}
}

//AddNode adds new node group to the existing cluster, cluster atleast have 1 node group already present
func (ctrl AWSController) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {

	//create a new node on the given cluster with the NodeSpec
	clusterName := req.ClusterName
	region := req.Region
	nodeSpec := req.NodeSpec

	session, err := NewSession(ctrl.config, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	cluster, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {

		ctrl.logger.Errorf("unable to get cluster '%s': %s", clusterName, err.Error())
		return nil, err
	}

	ctrl.logger.Infof("querying default nodes on cluster '%s' in region '%s'", clusterName, region)
	defaultNode, err := ctrl.getDefaultNode(ctx, client, clusterName, nodeSpec.Name)

	var newNodeGroupInput *eks.CreateNodegroupInput

	if err != nil {
		if errors.Is(err, ERR_NODEGROUP_EXIST) {
			ctrl.logger.Errorf("nodegroup '%s' already exist in cluster '%s'", nodeSpec.Name, clusterName)
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
	return &pb.NodeSpawnResponse{}, err
}

//DeleteCluster delete empty cluster, cluster should not have any nodegroup attached.
func (ctrl AWSController) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {

	clusterName := req.ClusterName
	region := req.Region

	session, err := NewSession(ctrl.config, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	deleteOut, err := client.DeleteClusterWithContext(ctx, &eks.DeleteClusterInput{
		Name: &clusterName,
	})

	if err != nil {
		ctrl.logger.Errorf("failed to delete cluster '%s': %s", clusterName, err.Error())
		return &pb.ClusterDeleteResponse{
			Error: err.Error(),
		}, err
	}

	ctrl.logger.Infof("requested cluster '%s' to be deleted, Status :%s. It might take some time, check AWS console for more.", clusterName, *deleteOut.Cluster.Status)

	return &pb.ClusterDeleteResponse{}, nil
}

func (ctrl AWSController) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	clusterName := req.ClusterName
	nodeName := req.NodeGroupName
	region := req.Region

	session, err := NewSession(ctrl.config, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	nodeDeleteOut, err := client.DeleteNodegroupWithContext(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodeName,
	})

	if err != nil {
		ctrl.logger.Errorf("failed to delete nodegroup '%s': %s", nodeName, err.Error())
		return &pb.NodeDeleteResponse{Error: err.Error()}, err
	}
	ctrl.logger.Infof("requested nodegroup '%s' to be deleted, Status %s. It might take some time, check AWS console for more.", nodeName, *nodeDeleteOut.Nodegroup.Status)
	return &pb.NodeDeleteResponse{}, nil
}

func (ctrl AWSController) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	return &pb.AddTokenResponse{}, nil
}

func (ctrl AWSController) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {

	region := req.Region
	clusterName := req.ClusterName

	session, err := NewSession(ctrl.config, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()
	ctrl.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
	cluster, err := getClusterSpec(ctx, client, clusterName)

	kubeConfig, err := session.getKubeConfig(cluster)
	if err != nil {
		ctrl.logger.Errorf("failed to get k8s %s", err.Error())
		return nil, err
	}
	return &pb.GetTokenResponse{
		Token:    kubeConfig.BearerToken,
		CaData:   string(kubeConfig.CAData),
		Endpoint: kubeConfig.Host,
	}, nil
}

func (ctrl AWSController) RegisterWithRancher(ctx context.Context, req *pb.RancherRegistrationRequest) (*pb.RancherRegistrationResponse, error) {
	return nil, nil
}
