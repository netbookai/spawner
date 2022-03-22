package azure

import (
	"context"
	"fmt"
	"net/http"

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

	aksClient, err := getAKSClient()
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	a.logger.Infow("fetching cluster information", "cluster", clusterName)
	clstr, err := aksClient.Get(ctx, groupName, clusterName)
	if err != nil {
		a.logger.Errorw("failed to get cluster ", "error", err)
		return nil, err
	}

	apc, err := getAgentPoolClient()
	if err != nil {
		a.logger.Errorw("failed to get agent pool client", "error", err)
		return nil, err
	}
	nodeName := req.NodeSpec.Name

	a.logger.Infow("cluster found, adding new node", "cluster", clusterName, "node", nodeName)

	nodeTags := labels.GetNodeLabel(req.NodeSpec)

	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/agent-pools/create-or-update
	future, err := apc.CreateOrUpdate(
		ctx,
		groupName,
		*clstr.Name,
		nodeName,
		containerservice.AgentPool{
			ManagedClusterAgentPoolProfileProperties: &containerservice.ManagedClusterAgentPoolProfileProperties{

				Count:               to.Int32Ptr(1),
				VMSize:              to.StringPtr(req.NodeSpec.Instance),
				Tags:                nodeTags,
				Mode:                containerservice.AgentPoolModeUser,
				OrchestratorVersion: &constants.AzureKubeVersion,
				OsDiskSizeGB:        &req.NodeSpec.DiskSize,
			},
		},
	)

	if err != nil {
		a.logger.Errorw("failed to add node", "error", err)
		return nil, errors.Wrapf(err, "failed to add node to the cluster")
	}

	a.logger.Infow("requested to add new node, waiting on completion")
	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		a.logger.Errorw("failed to add node", "error", err)
		return nil, errors.Wrapf(err, "failed to add node to the cluster")
	}

	return &proto.NodeSpawnResponse{}, nil
}

func (a *AzureController) deleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {

	groupName := config.Get().AzureResourceGroup
	apc, err := getAgentPoolClient()
	if err != nil {
		a.logger.Errorw("failed to get agent pool client", "error", err)
		return nil, err
	}

	cluster := req.GetClusterName()
	node := req.GetNodeGroupName()

	future, err := apc.Delete(ctx, groupName, cluster, node)

	if err != nil {
		a.logger.Errorw("failed to delete the node pool", "error", err)
		return nil, err
	}
	err = future.WaitForCompletionRef(ctx, apc.Client)

	if err != nil {
		a.logger.Errorw("failed to delete the node pool", "error", err)
		return nil, err
	}

	if future.Response().StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("request resource '%s' not found in cluster '%s'", node, cluster)
	}

	a.logger.Infow("delete node successfully", "status", future.Response().Status)
	return &proto.NodeDeleteResponse{}, nil
}
