package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func (a AzureController) addNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {

	config := config.Get()

	clusterName := req.ClusterName
	groupName := config.AzureResourceGroup
	region := req.Region

	clientID := config.AzureClientID
	clientSecret := config.AzureClientSecret

	aksClient, err := getAKSClient()
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	nodeTags := labels.GetNodeLabel(req.NodeSpec)
	future, err := aksClient.CreateOrUpdate(
		ctx,
		groupName,
		clusterName,
		containerservice.ManagedCluster{
			Name:     &clusterName,
			Location: &region,
			ManagedClusterProperties: &containerservice.ManagedClusterProperties{
				DNSPrefix: &clusterName,
				AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
					{

						Count:               to.Int32Ptr(1),
						Name:                to.StringPtr(req.NodeSpec.Name),
						VMSize:              to.StringPtr(req.NodeSpec.Instance),
						Tags:                nodeTags,
						Mode:                containerservice.AgentPoolModeUser,
						OrchestratorVersion: &constants.KubeVersion,
						OsDiskSizeGB:        &req.NodeSpec.DiskSize,
					},
				},
				ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
					ClientID: to.StringPtr(clientID),
					Secret:   to.StringPtr(clientSecret),
				},
			},
		},
	)

	if err != nil {
		a.logger.Errorw("failed to add node", "error", err)
		return nil, errors.Wrapf(err, "failed to add node to the cluster")
	}

	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		a.logger.Errorw("failed to add node", "error", err)
		return nil, errors.Wrapf(err, "failed to add node to the cluster")
	}

	return &proto.NodeSpawnResponse{}, nil
}
