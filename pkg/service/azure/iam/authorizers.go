package iam

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"

	"github.com/Azure/go-autorest/autorest/azure"
)

// GetResourceManagementAuthorizer gets an OAuthTokenAuthorizer for Azure Resource Manager
func GetResourceManagementAuthorizer(cred *system.AzureCredential) (autorest.Authorizer, error) {
	return getAuthorizerForResource(cred)
}

func getAuthorizerForResource(cred *system.AzureCredential) (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error
	environments, err := azure.EnvironmentFromName(config.Get().AzureCloudProvider)

	if err != nil {
		return nil, errors.Wrapf(err, "invalid azure cloud provider '%s'", config.Get().AzureCloudProvider)
	}

	oauthConfig, err := adal.NewOAuthConfig(environments.ActiveDirectoryEndpoint, cred.TenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig,
		cred.ClientID,
		cred.ClientSecret,
		environments.ResourceManagerEndpoint)

	if err != nil {
		return nil, err
	}

	a = autorest.NewBearerAuthorizer(token)
	return a, nil
}
