package aws

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

//arn:aws:iam::965734315247:oidc-provider
//
// oidc endpoint
const trustPolicTemplate = `{
			"Effect": "Allow",
			"Principal": {
				"Federated": "%s"
			},
			"Action": "sts:AssumeRoleWithWebIdentity",
			"Condition": {
				"StringEquals": {
					"%s:sub": "system:serviceaccount:%s:%s",
					"%s:aud": "sts.amazonaws.com"
				}
			}
		}`

func getTrustPolicyDocument(federatedPrefix, oidcUrl string) string {
	return fmt.Sprintf(trustPolicTemplate, federatedPrefix, oidcUrl, constants.OIDCNamespace, constants.OIDCServiceAccount, oidcUrl)
}

type partialOIDCConfig struct {
	JwksURI string `json:"jwks_uri"`
}

func getJwksURL(ctx context.Context, openidConfigURL string) (string, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", openidConfigURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var partialOIDCConfig partialOIDCConfig
	if err := json.Unmarshal(body, &partialOIDCConfig); err != nil {
		return "", err
	}

	return partialOIDCConfig.JwksURI, nil
}

// getThumbprint will get the root CA from TLS certificate chain for the FQDN of the JWKS URL.
func getThumbprint(ctx context.Context, jwksURL string) (string, error) {
	parsedURL, err := url.Parse(jwksURL)
	if err != nil {
		return "", err
	}
	hostname := parsedURL.Host
	if parsedURL.Port() == "" {
		hostname = net.JoinHostPort(hostname, "443")
	}

	url := "https://" + hostname
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	peerCerts := resp.TLS.PeerCertificates
	numCerts := len(peerCerts)
	if numCerts == 0 {
		return "", fmt.Errorf("could not find any peer certificates for URL %s", jwksURL)
	}

	// root CA certificate is the last one in the list
	root := peerCerts[numCerts-1]
	return sha1Hash(root.Raw), nil
}

// sha1Hash computes the SHA1 of the byte array and returns the hex encoding as a string.
func sha1Hash(data []byte) string {
	hasher := sha1.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}

// getOIDCConfigURL constructs the URL where you can retrieve the OIDC Config information for a given OIDC provider.
func getOIDCConfigURL(issuerURL string) (string, error) {
	parsedURL, err := url.Parse(issuerURL)
	if err != nil {
		return "", err
	}

	parsedURL.Path = path.Join(parsedURL.Path, ".well-known", "openid-configuration")
	openidConfigURL := parsedURL.String()
	return openidConfigURL, nil
}

// getOIDCThumbprint will retrieve the thumbprint of the root CA for the OIDC Provider identified by the issuer URL.
// This is done by first looking up the domain where the keys are provided, and then looking up the TLS certificate
// chain for that domain.
func getOIDCThumbprint(ctx context.Context, issuerURL string) (string, error) {

	openidConfigURL, err := getOIDCConfigURL(issuerURL)
	if err != nil {
		return "", errors.Wrap(err, "getOIDCConfigURL")
	}

	jwksURL, err := getJwksURL(ctx, openidConfigURL)
	if err != nil {
		return "", errors.Wrap(err, "getJwksURL")
	}

	return getThumbprint(ctx, jwksURL)
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

func (a *awsController) RegisterClusterOIDC(ctx context.Context, req *proto.RegisterClusterOIDCRequest) (*proto.RegisterClusterOIDCResponse, error) {
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

	if *cluster.Status != eks.ClusterStatusActive {
		a.logger.Info(ctx, "cluster is not active yet", "status", *cluster.Status)
		return nil, fmt.Errorf("cluster is not active, curent state: %s", *cluster.Status)
	}
	if cluster.Identity == nil {
		a.logger.Info(ctx, "cluster doesnt have identity", "identity", nil)
		return nil, errors.New("cluster identity is nil")
	}

	if cluster.Identity.Oidc == nil {
		a.logger.Info(ctx, "cluster doesnt have oidc identity", "identity.oidc", nil)
		return nil, errors.New("cluster oidc identity is nil")
	}

	if cluster.Identity.Oidc.Issuer == nil {
		a.logger.Info(ctx, "cluster oidc identity issuer is nil", "identity.oidc.issuer", nil)
		return nil, errors.New("cluster oidc identity issuer is nil")
	}

	issuer := *cluster.Identity.Oidc.Issuer
	if issuer == "" {
		a.logger.Info(ctx, "cluster oidc identity issuer is empty", "identity.oidc.issuer", "")
		return nil, errors.New("cluster oidc identity issuer is empty")
	}

	a.logger.Info(ctx, "cluster found", "issuer", issuer)

	iamClient := session.getIAMClient()
	accountId, err := session.getAccountId()
	if err != nil {
		a.logger.Error(ctx, "failed to get the account id", "error", err)
		return nil, errors.Wrap(err, "getAccountid")
	}

	clusterUrl := strings.Split(issuer, "https://")[1]
	providerArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountId, clusterUrl)
	//check if we already created a open id oidcProvider
	_, err = iamClient.GetOpenIDConnectProviderWithContext(ctx, &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: &providerArn,
	})

	if err != nil {
		a.logger.Warn(ctx, "failed to fetch open id provider", "error", err)

		a.logger.Info(ctx, "generating OIDC thumbprint ", "issuer", issuer)
		thumbprint, err := getOIDCThumbprint(ctx, issuer) //sample: "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
		if err != nil {
			a.logger.Error(ctx, "failed to get oidc thumbprint", "error", err, "issuer", issuer)
			return nil, errors.Wrap(err, "getOIDCThumbprint")
		}
		r, err := iamClient.CreateOpenIDConnectProviderWithContext(ctx, &iam.CreateOpenIDConnectProviderInput{
			ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
			ThumbprintList: []*string{&thumbprint},
			Url:            &issuer,
		})
		if err != nil {
			a.logger.Error(ctx, "failed to create open id connect provider", "error", err)
			return nil, errors.Wrap(err, "CreateOpenIDConnectProvider ")
		}
		providerArn = *r.OpenIDConnectProviderArn
	}

	role, err := iamClient.GetRoleWithContext(ctx, &iam.GetRoleInput{
		RoleName: &roleName,
	})
	if err != nil {
		a.logger.Error(ctx, "failed to get the oidc role", "error", err)
		return nil, errors.Wrap(err, "failed to get the role")
	}

	//get the current policy doc and append current cluster statement
	newPolicy, err := generateTrustPolicyDocument(*role.Role.AssumeRolePolicyDocument, providerArn, clusterUrl)
	if err != nil {
		a.logger.Error(ctx, "failed to generate trust policyDocument", "error", err)
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
	return &proto.RegisterClusterOIDCResponse{}, nil
}
