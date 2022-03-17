package iam

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"

	"github.com/Azure/go-autorest/autorest/azure"
)

var (
	armAuthorizer autorest.Authorizer
)

// GetResourceManagementAuthorizer gets an OAuthTokenAuthorizer for Azure Resource Manager
func GetResourceManagementAuthorizer() (autorest.Authorizer, error) {
	if armAuthorizer != nil {
		return armAuthorizer, nil
	}

	var a autorest.Authorizer
	var err error

	a, err = getAuthorizerForResource()

	if err == nil {
		// cache
		armAuthorizer = a
	} else {
		// clear cache
		armAuthorizer = nil
	}
	return armAuthorizer, err
}

func getAuthorizerForResource() (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error
	config := config.Get()
	environments, err := azure.EnvironmentFromName(config.AzureCloudProvider)

	if err != nil {
		return nil, errors.Wrapf(err, "invalid azure cloud provider '%s'", config.AzureCloudProvider)
	}

	oauthConfig, err := adal.NewOAuthConfig(
		environments.ActiveDirectoryEndpoint, config.AzureTenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig,
		config.AzureClientID,
		config.AzureClientSecret,
		environments.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	a = autorest.NewBearerAuthorizer(token)

	return a, err
}
