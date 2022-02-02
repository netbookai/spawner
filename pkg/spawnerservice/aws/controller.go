package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/rancher/common"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWS_CLUSTER_ROLE_NAME    = "netbook-AWS-ServiceRoleForEKS-BADBEEF2"
	AWS_NODE_GROUP_ROLE_NAME = "netbook-AWS-NodeGroupInstanceRole-CAFE1"
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

type AWSController struct {
	logger         *zap.SugaredLogger
	config         *config.Config
	ec2SessFactory func(region string) (awssession ec2iface.EC2API, err error)
}

func Ec2SessionFactory(region string) (awsSession ec2iface.EC2API, err error) {
	sess, err := CreateBaseSession(region)
	if err != nil {
		return nil, errors.Wrap(err, "Can't start AWS session")
	}

	awsSvc := ec2.New(sess)
	return awsSvc, err
}

func NewAWSController(logger *zap.SugaredLogger, config *config.Config) AWSController {
	return AWSController{
		logger:         logger,
		config:         config,
		ec2SessFactory: Ec2SessionFactory,
	}
}

func (svc AWSController) createRole(ctx context.Context, iamClient *iam.IAM, roleName string, description string, assumeRoleDoc string) (*iam.Role, error) {

	role, err := iamClient.GetRoleWithContext(ctx, &iam.GetRoleInput{
		RoleName: &roleName,
	})

	if err == nil {
		svc.logger.Debugf("role '%s' found, using the same", roleName)
		return role.Role, nil
	}
	//role not found, create it
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == iam.ErrCodeNoSuchEntityException {
		svc.logger.Warnf("failed to get role '%s', creating new role", roleName)
		//role does not exist, create one

		roleInput := &iam.CreateRoleInput{
			RoleName:                 &roleName,
			AssumeRolePolicyDocument: &assumeRoleDoc,
			Description:              &description,
			Tags: []*iam.Tag{{
				Key:   common.StrPtr(constants.CREATOR_LABEL),
				Value: common.StrPtr(constants.SPAWNER_SERVICE_LABEL),
			},
				{
					Key:   common.StrPtr("Name"),
					Value: &roleName,
				},
			},
		}

		roleOut, err := iamClient.CreateRoleWithContext(ctx, roleInput)
		if err != nil {
			svc.logger.Errorf("failed to query and create new role, %w", err)
			return nil, err
		}
		fmt.Println("role created", roleOut)

		return roleOut.Role, nil
	} else {
		return nil, err
	}
}

func (svc AWSController) attachPolicy(ctx context.Context, iamClient *iam.IAM, roleName string, policyARN string) error {
	//attach arn:aws:iam::aws:policy/AmazonEKSClusterPolicy

	attachPolicyInput := &iam.AttachRolePolicyInput{
		PolicyArn: &policyARN,
		RoleName:  &roleName,
	}

	_, err := iamClient.AttachRolePolicyWithContext(ctx, attachPolicyInput)
	return err
}

func (svc AWSController) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {

	var clusterName string
	if clusterName = req.ClusterName; len(clusterName) == 0 {
		clusterName = fmt.Sprintf("%s-%s", req.Provider, req.Region)
	}

	region := req.Region
	session, err := CreateBaseSession(region)

	if err != nil {
		return nil, err
	}
	client := eks.New(session)

	//TODO: check if cluster already exists with the name?

	//setup network

	var subnets []*string

	awsRegionNetworkStack, err := GetRegionWkspNetworkStack(region)
	if err != nil {
		svc.logger.Errorw("error getting network stack for region", "region", region, "error", err)
		return nil, err
	}

	if awsRegionNetworkStack.Vpc != nil && len(awsRegionNetworkStack.Subnets) > 0 {
		for _, subn := range awsRegionNetworkStack.Subnets {
			subnets = append(subnets, subn.SubnetId)
		}
		svc.logger.Infow("got network stack for region", "vpc", awsRegionNetworkStack.Vpc.VpcId, "subnets", subnets)
	} else {
		awsRegionNetworkStack, err = CreateRegionWkspNetworkStack(region)
		if err != nil {
			svc.logger.Errorw("error creating network stack for region with no clusters", "region", region, "error", err)
			svc.logger.Warnw("rolling back network stack changes as creation failed", "region", region)
			delErr := DeleteRegionWkspNetworkStack(region, *awsRegionNetworkStack)
			if delErr != nil {
				svc.logger.Errorw("error deleting network stack for region", "region", region, "error", delErr)
			}

			return nil, err
		}
		for _, subn := range awsRegionNetworkStack.Subnets {
			subnets = append(subnets, subn.SubnetId)
		}
		svc.logger.Infow("created network stack for region", "vpc", awsRegionNetworkStack.Vpc.VpcId, "subnets", subnets)
	}
	tags := map[string]*string{
		constants.CLUSTER_NAME_LABEL: &clusterName,
		constants.CREATOR_LABEL:      common.StrPtr(constants.SPAWNER_SERVICE_LABEL),
		constants.PROVISIONER_LABEL:  common.StrPtr(constants.RANCHER_LABEL)}

	for k, v := range req.Labels {
		tags[k] = &v
	}

	iamClient := iam.New(session)
	date := time.Now().Format("01-02-2006")
	roleName := fmt.Sprintf("%s-%s", AWS_CLUSTER_ROLE_NAME, date)

	eksRole, err := svc.createRole(ctx, iamClient, roleName, "eks cluster and service access role", EKS_ASSUME_ROLE_DOC)

	if err != nil {
		svc.logger.Errorf("failed to create role %w", err)
		return nil, err
	}
	err = svc.attachPolicy(ctx, iamClient, roleName, EKS_CLUSTER_POLICY_ARN)
	if err != nil {
		svc.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_CLUSTER_POLICY_ARN, roleName, err)
		return nil, err
	}

	err = svc.attachPolicy(ctx, iamClient, roleName, EKS_SERVICE_POLICY_ARN)
	if err != nil {
		svc.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_SERVICE_POLICY_ARN, roleName, err)
		return nil, err
	}
	clusterInput := &eks.CreateClusterInput{
		Name: &clusterName,
		ResourcesVpcConfig: &eks.VpcConfigRequest{
			SubnetIds:             subnets,
			EndpointPublicAccess:  common.BoolPtr(true),
			EndpointPrivateAccess: common.BoolPtr(false),
		},
		Tags:    tags,
		Version: common.StrPtr("1.20"),
		RoleArn: eksRole.Arn,
	}

	createClusterOutput, err := client.CreateClusterWithContext(ctx, clusterInput)
	if err != nil {
		svc.logger.Error("failed to create cluster", err)
	}
	fmt.Println(createClusterOutput)

	//create node group policy
	nodeRole, err := svc.createRole(ctx, iamClient, AWS_NODE_GROUP_ROLE_NAME, "node group instance policy role", EC2_ASSUME_ROLE_DOC)

	if err != nil {
		svc.logger.Errorf("failed to create node group role '%s' %w", AWS_NODE_GROUP_ROLE_NAME, err)
		return nil, err
	}

	err = svc.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_WORKER_NODE_POLICY_ARN)

	if err != nil {
		svc.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_WORKER_NODE_POLICY_ARN, AWS_NODE_GROUP_ROLE_NAME, err)
		return nil, err
	}

	err = svc.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_EC2_CONTAINER_RO_POLICY_ARN)

	if err != nil {
		svc.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_EC2_CONTAINER_RO_POLICY_ARN, AWS_NODE_GROUP_ROLE_NAME, err)
		return nil, err
	}

	err = svc.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_CNI_POLICY_ARN)

	if err != nil {
		svc.logger.Errorf("failed to attach policy '%s' to role '%s' %w", EKS_CNI_POLICY_ARN, AWS_NODE_GROUP_ROLE_NAME, err)
		return nil, err
	}

	//add node

	nodeSpec := req.Node
	diskSize := int64(nodeSpec.DiskSize)

	labels := make(map[string]*string)
	for k, v := range nodeSpec.Labels {
		labels[k] = &v
	}

	releaseVersion := ""

	newNodeGroup := &eks.CreateNodegroupInput{
		//Choose Amazon Linux 2 (AL2_x86_64) for Linux non-GPU instances, Amazon Linux 2 GPU Enabled (AL2_x86_64_GPU) for Linux GPU instances
		AmiType:        common.StrPtr("AL2_x86_64"),
		CapacityType:   common.StrPtr("ON_DEMAND"),
		NodeRole:       nodeRole.Arn,
		InstanceTypes:  []*string{&nodeSpec.Instance},
		ClusterName:    &clusterName,
		DiskSize:       &diskSize,
		NodegroupName:  &nodeSpec.Name,
		ReleaseVersion: &releaseVersion,
		Labels:         labels,
		Subnets:        subnets,
	}

	// wait till cluster becomes active

	eksClient := eks.New(session)
	errCount := 0
	for {

		fmt.Println("polling for cluster status...")
		svc.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
		resp, err := getClusterSpec(ctx, eksClient, clusterName)

		if err != nil {
			fmt.Printf("error fetching status %w", err)
			errCount += 1
		} else {

			fmt.Printf(" Status %s", *resp.Cluster.Status)
			if *resp.Cluster.Status == "ACTIVE" {
				break
			}
		}

		if errCount == 20 {
			//failed all the attempts, break it
			return nil, fmt.Errorf("timeout waiting for AWS to create a cluster")
		}
		time.Sleep(time.Minute)
	}

	out, err := client.CreateNodegroupWithContext(ctx, newNodeGroup)
	if err != nil {
		svc.logger.Errorf("failed to add a node '%s': %w", nodeSpec.Name, err)
		return nil, err
	}
	svc.logger.Debug("creating nodegroup ", out)
	return &pb.ClusterResponse{}, err
}

func getClusterSpec(ctx context.Context, client *eks.EKS, name string) (*eks.DescribeClusterOutput, error) {
	input := eks.DescribeClusterInput{
		Name: &name,
	}
	resp, err := client.DescribeClusterWithContext(ctx, &input)
	return resp, err
}

func (svc AWSController) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {

	response := &pb.ClusterSpec{}
	region := "us-west-2"
	clusterName := req.ClusterName
	session, err := CreateBaseSession(region)

	svc.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
	if err != nil {
		return nil, err
	}
	client := eks.New(session)

	resp, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		svc.logger.Error("failed to fetch cluster status", err)
		return nil, err
	}

	k8sClient, err := newClientset(session, resp.Cluster)
	if err != nil {
		svc.logger.Error(" Failed to create kube client ", err)
		return nil, err
	}
	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	response.Name = clusterName

	if err != nil {
		svc.logger.Error(" Failed to query node list ", err)
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

func (svc AWSController) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {

	//TODO: what does Scope mean here ?

	//get all clusters in given region
	region := req.Region
	session, err := CreateBaseSession(region)
	if err != nil {
		return nil, err
	}

	client := eks.New(session)

	//list cluster allows paginated query,
	listClutsreInput := &eks.ListClustersInput{}
	listClutsreOut, err := client.ListClustersWithContext(ctx, listClutsreInput)
	if err != nil {
		svc.logger.Error("failed to list clusters", err)
		return &pb.GetClustersResponse{}, err
	}

	resp := pb.GetClustersResponse{
		Clusters: [](*pb.ClusterSpec){},
	}

	for _, cluster := range listClutsreOut.Clusters {

		//clusterDetails, _ := getClusterSpec(ctx, client, *cluster)
		input := &eks.ListNodegroupsInput{ClusterName: cluster}
		nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
		if err != nil {
			svc.logger.Error("failed to fetch nodegroups")
		}

		nodes := []*pb.NodeSpec{}
		for _, cNodeGroup := range nodeGroupList.Nodegroups {
			input := &eks.DescribeNodegroupInput{
				NodegroupName: cNodeGroup,
				ClusterName:   cluster}
			nodeGroupDetails, err := client.DescribeNodegroupWithContext(ctx, input)

			if err != nil {
				svc.logger.Error("failed to fetch nodegroups details ", *cNodeGroup)
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

func (svc AWSController) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	//todo: Should we get this from the request ARGS ?
	region := req.Region
	clusterName := req.ClusterName
	session, err := CreateBaseSession(region)

	if err != nil {
		return nil, err
	}
	client := eks.New(session)

	svc.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
	resp, err := getClusterSpec(ctx, client, clusterName)

	if err != nil {
		svc.logger.Error("failed to fetch cluster status", err)
		return &pb.ClusterStatusResponse{
			Error: err.Error(),
		}, err
	}

	return &pb.ClusterStatusResponse{
		Status: *resp.Cluster.Status,
	}, err
}

func (svc AWSController) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {

	//create a new node on the given cluster with the NodeSpec
	clusterName := req.ClusterName
	region := "us-west-2" //req.Region
	nodeSpec := req.NodeSpec

	session, err := CreateBaseSession(region)
	if err != nil {
		return nil, err
	}
	client := eks.New(session)

	//resp, err := getClusterSpec(ctx, client, clusterName)

	//if err != nil {
	//	svc.logger.Error("failed to fetch cluster status", err)
	//	return nil, err
	//}

	input := &eks.ListNodegroupsInput{ClusterName: &clusterName}
	nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
	if err != nil {
		svc.logger.Error("failed to fetch nodegroups", err)
		return nil, err
	}

	for _, nodeGroup := range nodeGroupList.Nodegroups {
		if *nodeGroup == nodeSpec.Name {
			return nil, fmt.Errorf("nodegroup already exists with name %s", nodeSpec.Name)
		}
	}

	nodeDetailsinput := &eks.DescribeNodegroupInput{
		NodegroupName: nodeGroupList.Nodegroups[0],
		ClusterName:   &clusterName}
	nodeGroupDetails, err := client.DescribeNodegroupWithContext(ctx, nodeDetailsinput)

	if err != nil {
		return nil, err
	}

	diskSize := int64(nodeSpec.DiskSize)

	labels := map[string]*string{
		constants.CREATOR_LABEL:             common.StrPtr(constants.SPAWNER_SERVICE_LABEL),
		constants.PROVISIONER_LABEL:         common.StrPtr(constants.RANCHER_LABEL),
		constants.NODE_NAME_LABEL:           &nodeSpec.Name,
		constants.NODE_LABEL_SELECTOR_LABEL: &nodeSpec.Name,
		constants.INSTANCE_LABEL:            &nodeSpec.Instance,
		"type":                              common.StrPtr("nodegroup")}

	for k, v := range nodeGroupDetails.Nodegroup.Labels {
		labels[k] = v
	}

	for k, v := range nodeSpec.Labels {
		labels[k] = &v
	}

	newNodeGroup := &eks.CreateNodegroupInput{
		AmiType:        nodeGroupDetails.Nodegroup.AmiType,
		CapacityType:   nodeGroupDetails.Nodegroup.CapacityType,
		NodeRole:       nodeGroupDetails.Nodegroup.NodeRole,
		InstanceTypes:  []*string{&nodeSpec.Instance},
		ClusterName:    &clusterName,
		DiskSize:       &diskSize,
		NodegroupName:  &nodeSpec.Name,
		ReleaseVersion: nodeGroupDetails.Nodegroup.ReleaseVersion,
		Labels:         labels,
		Subnets:        nodeGroupDetails.Nodegroup.Subnets,
	}
	out, err := client.CreateNodegroupWithContext(ctx, newNodeGroup)
	if err != nil {
		svc.logger.Errorf("failed to add a node '%s': %w", nodeSpec.Name, err)
		return nil, err
	}
	svc.logger.Debug("creating nodegroup ", out)
	return &pb.NodeSpawnResponse{}, err
}

func (svc AWSController) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	return &pb.ClusterDeleteResponse{}, nil
}

func (svc AWSController) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	return &pb.NodeDeleteResponse{}, nil
}

func (svc AWSController) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	return &pb.AddTokenResponse{}, nil
}

func (svc AWSController) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	return &pb.GetTokenResponse{}, nil
}
