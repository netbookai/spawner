package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/costmanagement/mgmt/2019-11-01/costmanagement"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/azure/iam"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
)

func getAKSClient(c *system.AzureCredential) (*containerservice.ManagedClustersClient, error) {

	aksClient := containerservice.NewManagedClustersClient(c.SubscriptionID)
	auth, err := iam.GetResourceManagementAuthorizer(c)
	if err != nil {
		return nil, err
	}
	aksClient.Authorizer = auth
	aksClient.AddToUserAgent(labels.SpawnerLabel)
	aksClient.PollingDuration = time.Hour * 1
	aksClient.RetryAttempts = 1
	return &aksClient, nil
}

func getCostManagementClient(c *system.AzureCredential) (*costmanagement.QueryClient, error) {

	costmgmtClient := costmanagement.NewQueryClient(c.SubscriptionID)
	auth, err := iam.GetResourceManagementAuthorizer(c)
	if err != nil {
		return nil, err
	}
	costmgmtClient.Authorizer = auth
	costmgmtClient.RetryAttempts = 1
	costmgmtClient.AddToUserAgent(labels.SpawnerLabel)

	return &costmgmtClient, nil
}

func getAgentPoolClient(c *system.AzureCredential) (*containerservice.AgentPoolsClient, error) {

	agentClient := containerservice.NewAgentPoolsClient(c.SubscriptionID)
	auth, err := iam.GetResourceManagementAuthorizer(c)
	if err != nil {
		return nil, err
	}
	agentClient.Authorizer = auth
	agentClient.AddToUserAgent(labels.SpawnerLabel)
	agentClient.PollingDuration = time.Hour * 1
	return &agentClient, nil
}

func getDisksClient(c *system.AzureCredential) (*compute.DisksClient, error) {
	dc := compute.NewDisksClient(c.SubscriptionID)
	a, err := iam.GetResourceManagementAuthorizer(c)

	if err != nil {
		return nil, err
	}
	dc.Authorizer = a
	dc.AddToUserAgent(labels.SpawnerLabel)
	return &dc, nil
}

func getSnapshotClient(c *system.AzureCredential) (*compute.SnapshotsClient, error) {
	sc := compute.NewSnapshotsClient(c.SubscriptionID)
	a, err := iam.GetResourceManagementAuthorizer(c)

	if err != nil {
		return nil, err
	}
	sc.Authorizer = a
	sc.AddToUserAgent(labels.SpawnerLabel)
	return &sc, nil
}
