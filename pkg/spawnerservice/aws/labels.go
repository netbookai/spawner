package aws

import (
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func getNodeLabel(nodeSpec *proto.NodeSpec) map[string]*string {
	labels := map[string]*string{
		constants.CreatorLabel:           common.StrPtr(constants.SpawnerServiceLabel),
		constants.NodeNameLabel:          &nodeSpec.Name,
		constants.InstanceLabel:          &nodeSpec.Instance,
		constants.NodeLabelSelectorLabel: &nodeSpec.Name,
		"type":                           common.StrPtr("nodegroup")}

	for k, v := range nodeSpec.Labels {
		labels[k] = &v
	}

	return labels

}
