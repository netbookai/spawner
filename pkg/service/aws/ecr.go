package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (a *AWSController) GetElasticRegistryAuth(ctx context.Context, req *proto.GetElasticRegistryAuthRequest) (*proto.GetElasticRegistryAuthResponse, error) {

	//	region := req.Region

	session, err := NewSession(ctx, req.Region, req.GetAccountName())

	if err != nil {
		a.logger.Error(ctx, "can't start AWS session", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to get aws session")
	}

	//get account id
	stsClient := session.getSTSClient()

	callerIdentity, err := stsClient.GetCallerIdentity(nil)

	if err != nil {
		a.logger.Error(ctx, "failed to get identity", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to get callerIdentity")
	}

	account_id := callerIdentity.Account
	a.logger.Debug(ctx, "fetched accountId", "id", account_id)

	//get ecr token
	ecrClient := session.getEcrClient()

	res, err := ecrClient.GetAuthorizationTokenWithContext(ctx, &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{account_id},
	})

	if err != nil {
		a.logger.Error(ctx, "failed to get the auth token for the ecr", "account_id", *account_id)
		return nil, errors.Wrap(err, "failed to get the auth token for the ecr")
	}

	if len(res.AuthorizationData) == 0 {
		a.logger.Error(ctx, "failed to get the auth token for the ecr, got empty AuthorizationData list", "account_id", *account_id)
		return nil, errors.New("got authorization data is empty")
	}
	ad := res.AuthorizationData[0]

	a.logger.Debug(ctx, "got the auth token for the ecr", "endpoint", *ad.ProxyEndpoint, "account", *account_id)
	return &proto.GetElasticRegistryAuthResponse{
		Url:   *ad.ProxyEndpoint,
		Token: *ad.AuthorizationToken,
	}, nil
}
