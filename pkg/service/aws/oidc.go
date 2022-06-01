package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

//arn:aws:iam::965734315247:oidc-provider
//
// oidc endpoint
const trustPolicTemplate = `{
			"Effect": "Allow",
			"Principal": {
				"Federated": "%s/%s"
			},
			"Action": "sts:AssumeRoleWithWebIdentity",
			"Condition": {
				"StringEquals": {
					"%s:sub": "system:serviceaccount:oidc:awskube2iam",
					"%s:aud": "sts.amazonaws.com"
				}
			}
		}`

func getTrustPolicyDocument(federatedPrefix, oidcUrl string) string {
	return fmt.Sprintf(trustPolicTemplate, federatedPrefix, oidcUrl, oidcUrl, oidcUrl)
}

//generateTrustPolicyDocument read the current policy document as a map, create new policy using the template stringa and convert that to map
//finally append the new policy to Statement section of the current policy
func generateTrustPolicyDocument(currentPolicyDoc, oidcproviderArn, oidcIssuer string) (string, error) {
	decodedValue, err := url.QueryUnescape(currentPolicyDoc)
	if err != nil {
		return "", err
	}

	policyDoc := make(map[string]interface{})
	err = json.Unmarshal([]byte(decodedValue), &policyDoc)
	if err != nil {
		return "", err
	}

	newPolicy := getTrustPolicyDocument(oidcproviderArn, oidcIssuer)

	newDoc := make(map[string]interface{})
	err = json.Unmarshal([]byte(newPolicy), &newDoc)
	if err != nil {
		return "", err
	}
	policyDoc["Statement"] = append(policyDoc["Statement"].([]interface{}), newDoc)

	b, err := json.Marshal(policyDoc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// fetch OIDC from the cluster and attach it to role policy

func (a *awsController) ConnectClusterOIDCToTrustPolicy(ctx context.Context, req *proto.ConnectClusterOIDCToTrustPolicyRequest) (*proto.ConnectClusterOIDCToTrustPolicyResponse, error) {
	//get the oidc endpoint from the cluster spec

	roleName := config.Get().OpenIDRole
	region := req.Region
	accountName := req.AccountName
	session, err := NewSession(ctx, region, accountName)

	if err != nil {
		return nil, err
	}
	eksClient := session.getEksClient()

	cluster, err := getClusterSpec(ctx, eksClient, req.ClusterName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the cluster")
	}
	issuer := *cluster.Identity.Oidc.Issuer

	a.logger.Info(ctx, "cluster found", "issuer", issuer)

	iamClient := session.getIAMClient()

	//	providerArn := "arn:aws:iam::965734315247:oidc-provider"
	//TODO: where is this coming from ? @mani
	thumbprint := "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
	r, err := iamClient.CreateOpenIDConnectProviderWithContext(ctx, &iam.CreateOpenIDConnectProviderInput{

		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
		ThumbprintList: []*string{&thumbprint},
		Url:            &issuer,
	})
	if err != nil {
		a.logger.Error(ctx, "failed to create open id connect provider", "error", err)
		return nil, errors.Wrap(err, "CreateOpenIDConnectProvider ")
	}
	providerArn := *r.OpenIDConnectProviderArn

	role, err := iamClient.GetRoleWithContext(ctx, &iam.GetRoleInput{
		RoleName: &roleName,
	})
	if err != nil {
		a.logger.Error(ctx, "failed to get the oidc role", "error", err)
		return nil, errors.Wrap(err, "failed to get the role")
	}

	//get the current policy doc and append current cluster statement
	newPolicy, err := generateTrustPolicyDocument(*role.Role.AssumeRolePolicyDocument, providerArn, issuer)
	if err != nil {
		a.logger.Error(ctx, "generateTrustPolicyDocument ")
		return nil, errors.Wrap(err, "failed to generate the new policy doc")
	}

	_, err = iamClient.UpdateAssumeRolePolicyWithContext(ctx, &iam.UpdateAssumeRolePolicyInput{
		PolicyDocument: &newPolicy,
		RoleName:       &roleName,
	})
	if err != nil {
		a.logger.Error(ctx, "update assume role policy failed", "error", err)
		return nil, errors.Wrap(err, "UpdateAssumeRolePolicy ")
	}

	a.logger.Info(ctx, "updated role policy document")
	return &proto.ConnectClusterOIDCToTrustPolicyResponse{}, nil
}
