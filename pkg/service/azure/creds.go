package azure

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
)

func getCredentials(ctx context.Context, account string) (*system.AzureCredential, error) {
	env := config.Get().Env

	if env == "local" {
		conf := config.Get()
		return &system.AzureCredential{
			SubscriptionID: conf.AzureSubscriptionID,
			TenantID:       conf.AzureTenantID,
			ClientID:       conf.AzureClientID,
			ClientSecret:   conf.AzureClientSecret,
			ResourceGroup:  conf.AzureResourceGroup,
			Name:           account,
		}, nil
	} else {
		c, err := system.GetCredentials(ctx, config.Get().SecretHostRegion, account, constants.AzureLabel)
		if err != nil {
			return nil, errors.Wrap(err, "getCredentials")
		}
		return c.GetAzure(), nil
	}
}
