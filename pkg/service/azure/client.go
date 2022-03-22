package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
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
	aksClient.RetryAttempts = 1
	return &aksClient, nil
}

func getAgentPoolClient() (*containerservice.AgentPoolsClient, error) {
	config := config.Get()

	agentClient := containerservice.NewAgentPoolsClient(config.AzureSubscriptionID)
	auth, err := iam.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, err
	}
	agentClient.Authorizer = auth
	agentClient.AddToUserAgent(constants.SpawnerServiceLabel)
	agentClient.PollingDuration = time.Hour * 1
	return &agentClient, nil
}

func getDisksClient() (*compute.DisksClient, error) {
	config := config.Get()
	disksClient := compute.NewDisksClient(config.AzureSubscriptionID)
	a, err := iam.GetResourceManagementAuthorizer()

	if err != nil {
		return nil, err
	}
	disksClient.Authorizer = a
	disksClient.AddToUserAgent(constants.SpawnerServiceLabel)
	return &disksClient, nil
}
