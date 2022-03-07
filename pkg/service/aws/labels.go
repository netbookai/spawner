package aws

import (
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
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
		v := v
		labels[k] = &v
	}

	return labels

}
