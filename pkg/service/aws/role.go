package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
)

//createRoleOrGetExisting creates a role if it does not exist
func (svc AWSController) createRoleOrGetExisting(ctx context.Context, iamClient *iam.IAM, roleName string, description string, assumeRoleDoc string) (*iam.Role, bool, error) {

	role, err := iamClient.GetRoleWithContext(ctx, &iam.GetRoleInput{
		RoleName: &roleName,
	})

	if err == nil {
		svc.logger.Info(ctx, "role found, using the same", "role", roleName)
		return role.Role, false, nil

	}

	key := func(k labels.Label) *string {
		return aws.String(k.Key())
	}

	//role not found, create it
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == iam.ErrCodeNoSuchEntityException {
		svc.logger.Warn(ctx, "failed to get role, creating new role", "role", roleName)
		//role does not exist, create one

		roleInput := &iam.CreateRoleInput{
			RoleName:                 &roleName,
			AssumeRolePolicyDocument: &assumeRoleDoc,
			Description:              &description,
			Tags: []*iam.Tag{
				{
					Key:   key(labels.CreatorLabel),
					Value: aws.String(labels.SpawnerLabel),
				},
				{
					Key:   key(labels.NameLabel),
					Value: &roleName,
				},
				{
					Key:   key(labels.Scope),
					Value: aws.String( /*(internal)aws.*/ labels.ScopeTag()),
				},
			},
		}

		roleOut, err := iamClient.CreateRoleWithContext(ctx, roleInput)
		if err != nil {
			svc.logger.Error(ctx, "failed to query and create new role", "error", err)
			return nil, false, err
		}
		svc.logger.Info(ctx, "role created", "role", *roleOut.Role.RoleName)

		return roleOut.Role, true, nil
	} else {
		return nil, false, err
	}
}

//attachPolicy attaches policy to given role
func (svc AWSController) attachPolicy(ctx context.Context, iamClient *iam.IAM, roleName string, policyARN string) error {
	//attach arn:aws:iam::aws:policy/AmazonEKSClusterPolicy

	attachPolicyInput := &iam.AttachRolePolicyInput{
		PolicyArn: &policyARN,
		RoleName:  &roleName,
	}

	_, err := iamClient.AttachRolePolicyWithContext(ctx, attachPolicyInput)
	return err
}
