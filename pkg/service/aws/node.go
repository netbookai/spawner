package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

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

//getDefaultNode Get any existing node from the cluster as default node
//if node with `newNode` exist return error
func (ctrl AWSController) getDefaultNode(ctx context.Context, client *eks.EKS, clusterName, nodeName string) (*eks.Nodegroup, error) {

	input := &eks.ListNodegroupsInput{ClusterName: &clusterName}
	nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to fetch nodegroups: %s", err.Error())
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

func getInstance(nodeSpec *proto.NodeSpec) (string, []*string, error) {

	capacityType := eks.CapacityTypesOnDemand
	instanceTypes := []*string{}

	if nodeSpec.CapacityType == proto.CapacityType_SPOT {
		capacityType = eks.CapacityTypesSpot
		for _, i := range nodeSpec.SpotInstances {
			instanceTypes = append(instanceTypes, &i)
		}
	} else {
		instance := ""
		if nodeSpec.MachineType != "" {
			instance = common.GetInstance(constants.AwsLabel, nodeSpec.MachineType)
		}

		//if user has specified the Instance, we will override previous ask
		if nodeSpec.Instance != "" {
			instance = nodeSpec.Instance
		}

		if instance == "" {
			return "", nil, errors.New(constants.InvalidInstanceOrMachineType)
		}
		instanceTypes = append(instanceTypes, &instance)
	}
	return capacityType, instanceTypes, nil
}

//buildNodegroupInput build a new node group request
func (a *AWSController) buildNodegroupInput(ctx context.Context, session *Session, clusterName *string, nodeSpec *proto.NodeSpec, subnetIds []*string, nodeRoleArn *string) (*eks.CreateNodegroupInput, error) {

	diskSize := int64(nodeSpec.DiskSize)

	labels := labels.GetNodeLabel(nodeSpec)

	count := int64(1)
	if nodeSpec.Count != 0 {
		count = nodeSpec.Count
	}

	capacityType, instanceTypes, err := getInstance(nodeSpec)

	if err != nil {
		return nil, err
	}
	gpuEnabled := nodeSpec.GpuEnabled

	if common.IsGPU(nodeSpec.MachineType) {
		gpuEnabled = true
	}
	amiType := ""
	//Choose Amazon Linux 2 (AL2_x86_64) for Linux non-GPU instances, Amazon Linux 2 GPU Enabled (AL2_x86_64_GPU) for Linux GPU instances
	if gpuEnabled {
		a.logger.Info(ctx, "requested gpu node", "name", nodeSpec.Name, "instance ", instanceTypes, "machine_type", nodeSpec.MachineType)
		amiType = "AL2_x86_64_GPU"
	} else {
		amiType = "AL2_x86_64"
	}
	a.logger.Debug(ctx, "building node group input", "name", nodeSpec.Name, "instance ", instanceTypes, "machine_type", nodeSpec.MachineType)

	return &eks.CreateNodegroupInput{
		AmiType:       &amiType,
		CapacityType:  &capacityType,
		NodeRole:      nodeRoleArn,
		InstanceTypes: instanceTypes,
		ClusterName:   clusterName,
		DiskSize:      &diskSize,
		NodegroupName: &nodeSpec.Name,
		Labels:        labels,
		Subnets:       subnetIds,
		ScalingConfig: &eks.NodegroupScalingConfig{
			DesiredSize: &count,
			MinSize:     &count,
			MaxSize:     &count,
		},
		Tags: labels,
	}, nil
}

func (ctrl AWSController) getNewNodeGroupSpecFromCluster(ctx context.Context, session *Session, cluster *eks.Cluster, nodeSpec *proto.NodeSpec) (*eks.CreateNodegroupInput, error) {

	iamClient := session.getIAMClient()

	roleName := AWS_NODE_GROUP_ROLE_NAME
	nodeRole, newRole, err := ctrl.createRoleOrGetExisting(ctx, iamClient, roleName, "node group instance policy role", EC2_ASSUME_ROLE_DOC)

	if err != nil {
		ctrl.logger.Error(ctx, "failed to create node group role '%s' %w", AWS_NODE_GROUP_ROLE_NAME, err)
		return nil, err
	}

	if newRole {

		err = ctrl.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_WORKER_NODE_POLICY_ARN)

		if err != nil {
			ctrl.logger.Error(ctx, "failed to attach policy to role", "policy", EKS_WORKER_NODE_POLICY_ARN, "node-role", AWS_NODE_GROUP_ROLE_NAME, "error", err)
			return nil, err
		}

		err = ctrl.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_EC2_CONTAINER_RO_POLICY_ARN)

		if err != nil {
			ctrl.logger.Error(ctx, "failed to attach policy to role", "policy", EKS_EC2_CONTAINER_RO_POLICY_ARN, "node-role", AWS_NODE_GROUP_ROLE_NAME, "error", err)
			return nil, err
		}

		err = ctrl.attachPolicy(ctx, iamClient, *nodeRole.RoleName, EKS_CNI_POLICY_ARN)

		if err != nil {
			ctrl.logger.Error(ctx, "failed to attach policy to role", "policy", EKS_CNI_POLICY_ARN, "node-role", AWS_NODE_GROUP_ROLE_NAME, "error", err)
			return nil, err
		}
	}

	input, err := ctrl.buildNodegroupInput(ctx, session, cluster.Name, nodeSpec, cluster.ResourcesVpcConfig.SubnetIds, nodeRole.Arn)
	if err != nil {
		return nil, errors.Wrap(err, "getNewNodeGroupSpecFromCluster:")
	}
	return input, nil

}

func (ctrl AWSController) getNodeSpecFromDefault(ctx context.Context, session *Session, defaultNode *eks.Nodegroup, clusterName string, nodeSpec *proto.NodeSpec) (*eks.CreateNodegroupInput, error) {

	input, err := ctrl.buildNodegroupInput(ctx, session, &clusterName, nodeSpec, defaultNode.Subnets, defaultNode.NodeRole)
	if err != nil {
		return nil, errors.Wrap(err, "getNodeSpecFromDefault")
	}
	return input, nil
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

		ctrl.logger.Error(ctx, "unable to get cluster, spec", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	ctrl.logger.Info(ctx, "querying default nodes on cluster '%s' in region '%s'", clusterName, region)
	defaultNode, err := ctrl.getDefaultNode(ctx, client, clusterName, nodeSpec.Name)

	var newNodeGroupInput *eks.CreateNodegroupInput

	if err != nil {
		if errors.Is(err, ERR_NODEGROUP_EXIST) {
			return nil, err
		}

		if errors.Is(err, ERR_NO_NODEGROUP) {
			//no node group present,
			ctrl.logger.Info(ctx, "default nodegroup not found in cluster '%s', creating NodegroupRequest from cluster config ", clusterName)
			newNodeGroupInput, err = ctrl.getNewNodeGroupSpecFromCluster(ctx, session, cluster, nodeSpec)
			if err != nil {
				return nil, err
			}
		}
	} else {
		ctrl.logger.Info(ctx, "found default nodegroup '%s' in cluster '%s', creating NodegroupRequest from default node config", *defaultNode.NodegroupName, clusterName)
		newNodeGroupInput, err = ctrl.getNodeSpecFromDefault(ctx, session, defaultNode, clusterName, nodeSpec)
		if err != nil {
			return nil, err
		}
	}

	out, err := client.CreateNodegroupWithContext(ctx, newNodeGroupInput)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to add a node '%s': %w", nodeSpec.Name, err)
		return nil, err
	}
	ctrl.logger.Info(ctx, "creating nodegroup '%s' on cluster '%s', Status : %s, it might take some time. Please check AWS console.", nodeSpec.Name, clusterName, *out.Nodegroup.Status)
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
			ctrl.logger.Error(ctx, "error when deleting nodegroup", "nodegroup", nodeGroupName)
			return err
		}
	}
	ctrl.logger.Info(ctx, "all attached nodegroups are being deleted", "cluster", clusterName)
	return nil
}

//waitForAllNodegroupsDeletion wait until all attached node groups in the clusters are deleted.
//Wait until all nodes are deleted or  configrNodeDeletionTimeout, whichever is earlier
func (ctrl AWSController) waitForAllNodegroupsDeletion(ctx context.Context, client *eks.EKS, clusterName string) error {
	input := &eks.ListNodegroupsInput{ClusterName: &clusterName}
	nodeGroupList, err := client.ListNodegroupsWithContext(ctx, input)
	if err != nil {
		return errors.Wrap(err, "waitForAllNodegroupsDeletion")
	}

	if len(nodeGroupList.Nodegroups) == 0 {
		return nil
	}

	//+1 to give some room for worst case, all delete returned error and context timeout.
	errChan := make(chan error, len(nodeGroupList.Nodegroups)+1)
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Get().NodeDeletionTimeout))
	defer cancel()

	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	for _, nodeGroupName := range nodeGroupList.Nodegroups {

		wg.Add(1)
		go func(nodeGroupName string, errChan chan<- error) {
			waitErr := client.WaitUntilNodegroupDeletedWithContext(ctx, &eks.DescribeNodegroupInput{
				ClusterName:   &clusterName,
				NodegroupName: &nodeGroupName,
			})

			//we will ignore context cancelled error to avoid duplicate
			if waitErr != nil && errors.Is(waitErr, context.Canceled) {
				errChan <- waitErr
			}
			wg.Done()

		}(*nodeGroupName, errChan)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		//any error send it over err chan and will deal with it at the end.
		errChan <- ctx.Err()
		break
		// waiting for all nodegroup delete done
	case <-done:
		break
	}

	close(errChan)
	var aggErr error
	for e := range errChan {
		//Wrap will return err when the err is nil and it will never be set.
		if aggErr == nil {
			aggErr = e
		} else {
			aggErr = errors.Wrap(aggErr, e.Error())
		}
	}
	if aggErr != nil {
		return errors.Wrap(aggErr, "waitForAllNodegroupsDeletion: couldn't wait on all node deletion")
	}
	return nil
}

func (ctrl AWSController) deleteNode(ctx context.Context, client *eks.EKS, cluster, node string) error {
	nodeDeleteOut, err := client.DeleteNodegroupWithContext(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   &cluster,
		NodegroupName: &node,
	})

	if err != nil {
		return err

	}
	ctrl.logger.Info(ctx, "requested nodegroup to be deleted. It might take some time, check AWS console for more.", "nodegroup", node, "status", *nodeDeleteOut.Nodegroup.Status)
	return nil
}

//DeleteNode delete nodes attched to cluster which is created by spawner
func (ctrl AWSController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	clusterName := req.ClusterName
	nodeName := req.NodeGroupName
	region := req.Region

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()

	nodeGroup, err := client.DescribeNodegroupWithContext(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodeName,
	})

	if err != nil {
		ctrl.logger.Error(ctx, "failed to get nodegroup details", "error", err)
		return nil, err
	}

	if scope, ok := nodeGroup.Nodegroup.Tags[constants.Scope]; !ok || *scope != labels.ScopeTag() {
		ctrl.logger.Error(ctx, "nodegroup is not available in scope", "scope", labels.ScopeTag())
		return nil, fmt.Errorf("nodegroup '%s' not available in scope '%s'", nodeName, labels.ScopeTag())
	}

	err = ctrl.deleteNode(ctx, client, clusterName, nodeName)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to delete nodegroup", "nodename", nodeName)
		return &proto.NodeDeleteResponse{Error: err.Error()}, err
	}

	return &proto.NodeDeleteResponse{}, nil
}
