package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/eks"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
)

func (svc AWSController) createClusterInternal(ctx context.Context, session *Session, clusterName string, req *pb.ClusterRequest) (*eks.Cluster, error) {

	var subnetIds []*string
	region := session.Region

	awsRegionNetworkStack, err := GetRegionWkspNetworkStack(session)
	if err != nil {
		svc.logger.Errorw("error getting network stack for region", "region", region, "error", err)
		return nil, err
	}

	if awsRegionNetworkStack.Vpc != nil && len(awsRegionNetworkStack.Subnets) > 0 {
		for _, subn := range awsRegionNetworkStack.Subnets {
			subnetIds = append(subnetIds, subn.SubnetId)
		}
		svc.logger.Infow("got network stack for region", "vpc", awsRegionNetworkStack.Vpc.VpcId, "subnets", subnetIds)
	} else {
		awsRegionNetworkStack, err = CreateRegionWkspNetworkStack(session)
		if err != nil {
			svc.logger.Errorw("error creating network stack for region with no clusters", "region", region, "error", err)
			svc.logger.Warnw("rolling back network stack changes as creation failed", "region", region)
			delErr := DeleteRegionWkspNetworkStack(session, *awsRegionNetworkStack)
			if delErr != nil {
				svc.logger.Errorw("error deleting network stack for region", "region", region, "error", delErr)
			}

			return nil, err
		}
		for _, subn := range awsRegionNetworkStack.Subnets {
			subnetIds = append(subnetIds, subn.SubnetId)
		}
		svc.logger.Infow("created network stack for region", "vpc", awsRegionNetworkStack.Vpc.VpcId, "subnets", subnetIds)
	}
	tags := map[string]*string{
		constants.ClusterNameLabel: &clusterName,
		constants.CreatorLabel:     common.StrPtr(constants.SpawnerServiceLabel),
	}

	for k, v := range req.Labels {
		tags[k] = &v
	}

	iamClient := session.getIAMClient()
	roleName := AWS_CLUSTER_ROLE_NAME

	eksRole, newRole, err := svc.createRoleOrGetExisting(ctx, iamClient, roleName, "eks cluster and service access role", EKS_ASSUME_ROLE_DOC)

	if err != nil {
		svc.logger.Errorf("failed to create role %w", err)
		return nil, err
	}

	if newRole {
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
	}

	clusterInput := &eks.CreateClusterInput{
		Name: &clusterName,
		ResourcesVpcConfig: &eks.VpcConfigRequest{
			SubnetIds:             subnetIds,
			EndpointPublicAccess:  common.BoolPtr(true),
			EndpointPrivateAccess: common.BoolPtr(false),
		},
		Tags:    tags,
		Version: common.StrPtr("1.20"),
		RoleArn: eksRole.Arn,
	}

	client := session.getEksClient()
	createClusterOutput, err := client.CreateClusterWithContext(ctx, clusterInput)
	if err != nil {
		svc.logger.Errorf("failed to create cluster %s", err.Error())
	}

	return createClusterOutput.Cluster, nil

}
