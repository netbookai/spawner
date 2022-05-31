package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func asEcrTags(label map[string]string) []*ecr.Tag {
	tags := []*ecr.Tag{}
	for k, v := range label {
		key := k
		val := v
		tags = append(tags, &ecr.Tag{
			Key:   &key,
			Value: &val,
		})
	}

	//add spawner default tags
	for k, v := range labels.DefaultTags() {
		key := k
		val := v
		tags = append(tags, &ecr.Tag{
			Key:   &key,
			Value: val,
		})
	}

	return tags
}

func (a *AWSController) GetContainerRegistryAuth(ctx context.Context, req *proto.GetContainerRegistryAuthRequest) (*proto.GetContainerRegistryAuthResponse, error) {

	session, err := NewSession(ctx, req.Region, req.GetAccountName())

	if err != nil {
		a.logger.Error(ctx, "can't start AWS session", "error", err)
		return nil, errors.Wrap(err, "CreateContainerRegistryRepo: failed to get aws session")
	}

	//get account id
	account_id, err := session.getAccountId()

	if err != nil {
		a.logger.Error(ctx, "failed to get account id", "error", err)
		return nil, errors.Wrap(err, "CreateContainerRegistryRepo: failed to get accoutn id")
	}

	a.logger.Debug(ctx, "fetched accountId", "id", account_id)

	//get ecr token
	ecrClient := session.getEcrClient()

	// Doc : https://docs.aws.amazon.com/sdk-for-go/api/service/ecr/#ECR.GetAuthorizationToken
	res, err := ecrClient.GetAuthorizationTokenWithContext(ctx, &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{&account_id},
	})

	if err != nil {
		a.logger.Error(ctx, "failed to get the auth token for the ecr", "account_id", account_id)
		return nil, errors.Wrap(err, "failed to get the auth token for the ecr")
	}

	if len(res.AuthorizationData) == 0 {
		a.logger.Error(ctx, "failed to get the auth token for the ecr, got empty AuthorizationData list", "account_id", account_id)
		return nil, errors.New("got authorization data is empty")
	}
	ad := res.AuthorizationData[0]

	a.logger.Debug(ctx, "got the auth token for the ecr", "endpoint", *ad.ProxyEndpoint, "account", account_id)
	return &proto.GetContainerRegistryAuthResponse{
		Url:   *ad.ProxyEndpoint,
		Token: *ad.AuthorizationToken,
	}, nil
}

func (a *AWSController) CreateContainerRegistryRepo(ctx context.Context, req *proto.CreateContainerRegistryRepoRequest) (*proto.CreateContainerRegistryRepoResponse, error) {

	repoName := req.Name
	session, err := NewSession(ctx, req.Region, req.GetAccountName())

	if err != nil {
		a.logger.Error(ctx, "can't start AWS session", "error", err)
		return nil, errors.Wrap(err, "CreateContainerRegistryRepo: failed to get aws session")
	}

	//get account id
	account_id, err := session.getAccountId()

	if err != nil {
		a.logger.Error(ctx, "failed to get account id", "error", err)
		return nil, errors.Wrap(err, "CreateContainerRegistryRepo: failed to get accoutn id")
	}

	a.logger.Debug(ctx, "fetched accountId", "id", account_id)

	//get ecr token
	ecrClient := session.getEcrClient()

	tags := asEcrTags(req.Tags)

	// Doc : https://docs.aws.amazon.com/sdk-for-go/api/service/ecr/#ECR.CreateRepository
	res, err := ecrClient.CreateRepositoryWithContext(ctx, &ecr.CreateRepositoryInput{
		RegistryId:     &account_id,
		RepositoryName: &repoName,
		Tags:           tags,
	})

	if err != nil {
		a.logger.Error(ctx, "failed to create the registry repo", "error", err)
		return nil, errors.Wrap(err, "CreateContainerRegistryRepo : ")
	}
	return &proto.CreateContainerRegistryRepoResponse{
		RegistryId: *res.Repository.RegistryId,
		Url:        *res.Repository.RepositoryUri,
	}, nil
}
