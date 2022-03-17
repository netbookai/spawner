package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

const testpubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDL67TCv+MyUnT0gHUl2xpJF56TjCkcTKkXUjhIaUDY/gv/bFm5pVbvrHovKV/W2MrI5e9Ix2iQIiityWVABFEFWe7m0yx3ds49ZkM3kIflsqmPeywCcN8V2bMsiVwyrLBsboeRcbQyJJIrsb8A0mj3ooWFfT44I42YVCg4FOTsB+wmlthawBlMGKzZb8ITUMaN0VCtXfIslg6ptQHtficL/N1HW7FSXXiZPJaRi3kuCH18e/wCkP4eomWMZ6MQC1CIwGIkfh9K4pfuppfZ9HG+jyw0ha0LZ6utDbEULMPAtvgUZXB7+1vk1NTwi78p558Dk6fxWGRVgSQu7Qk4yddZ nishanth@nishanth-Legion-5-15ACH6"

func (az *AzureController) createAKSCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {

	//	_, err := os.Stat(sshPublicKeyPath)
	//	if err == nil {
	//		sshBytes, err := ioutil.ReadFile(sshPublicKeyPath)
	//		if err != nil {
	//			log.Fatalf("failed to read SSH key data: %v", err)
	//		}
	//		sshKeyData = string(sshBytes)
	//	} else {
	//	}
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

	az.logger.Infow("creating cluster in AKS", "name", clusterName, "resource-group", groupName)
	tags := labels.DefaultTags()
	for k, v := range req.Labels {
		v := v
		tags[k] = &v

	}
	nodeTags := labels.GetNodeLabel(req.Node)
	fmt.Println("craeting cluster")

	future, err := aksClient.CreateOrUpdate(
		ctx,
		groupName,
		clusterName,
		containerservice.ManagedCluster{
			Tags:     tags,
			Name:     &clusterName,
			Location: &region,
			ManagedClusterProperties: &containerservice.ManagedClusterProperties{
				DNSPrefix: &clusterName,
				AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
					{
						Count:               to.Int32Ptr(1),
						Name:                to.StringPtr("defaultagent"),
						VMSize:              to.StringPtr(req.Node.Instance),
						Tags:                nodeTags,
						Mode:                containerservice.AgentPoolModeSystem,
						OrchestratorVersion: &constants.KubeVersion,
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
		az.logger.Errorw("failed to create a AKS cluster", "error", err)
		return nil, fmt.Errorf("cannot create AKS cluster: %v", err)
	}

	az.logger.Infow("waiting on the future completion")
	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		az.logger.Errorw("failed to get the future response", "error", err)
		return nil, fmt.Errorf("cannot get the AKS cluster create or update future response: %v", err)
	}

	//return future.Result(aksClient)

	return &proto.ClusterResponse{ClusterName: clusterName}, nil
}

func (az *AzureController) getCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	clusterName := req.ClusterName
	config := config.Get()
	groupName := config.AzureResourceGroup
	aksClient, err := getAKSClient()
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	az.logger.Infow("fetching cluster information", "cluster", clusterName)
	clstr, err := aksClient.Get(ctx, groupName, clusterName)
	if err != nil {
		az.logger.Errorw("failed to get cluster ", "error", err)
		return nil, err
	}

	state := constants.Inactive

	node := (*clstr.AgentPoolProfiles)[0]

	if node.PowerState.Code == containerservice.CodeRunning {
		state = constants.Active
	}

	return &proto.ClusterSpec{
		Name: clusterName,
		NodeSpec: []*proto.NodeSpec{{
			Name:     *node.Name,
			Instance: *node.VMSize,
			Labels:   aws.StringValueMap(node.Tags),
			DiskSize: *node.OsDiskSizeGB,
			State:    state,
		}},
	}, nil
}

func (az *AzureController) deleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {

	aksClient, err := getAKSClient()
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	clusterName := req.ClusterName
	config := config.Get()
	groupName := config.AzureResourceGroup

	future, err := aksClient.Delete(ctx, groupName, clusterName)

	if err != nil {
		az.logger.Errorw("failed to delete the cluster ", "error", err, "cluster", clusterName)
		return nil, err
	}

	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		az.logger.Errorw("failed to get the future response", "error", err)
		return nil, fmt.Errorf("cannot get the AKS cluster create or update future response: %v", err)
	}
	az.logger.Infow("cluster deleted successfully", "cluster", clusterName, "response", future.Status())

	return &proto.ClusterDeleteResponse{}, nil

}
