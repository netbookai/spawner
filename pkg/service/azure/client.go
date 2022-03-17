package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/azure/iam"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
)

func getAKSClient() (*containerservice.ManagedClustersClient, error) {
	config := config.Get()

	aksClient := containerservice.NewManagedClustersClient(config.AzureSubscriptionID)
	auth, err := iam.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, err
	}
	aksClient.Authorizer = auth
	aksClient.AddToUserAgent(constants.SpawnerServiceLabel)
	aksClient.PollingDuration = time.Hour * 1
	return &aksClient, nil
}
