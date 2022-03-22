package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func merge(maps ...map[string]*string) map[string]*string {
	m := make(map[string]*string)

	for _, _m := range maps {
		for k, v := range _m {
			m[k] = v
		}
	}
	return m
}

func getNodeLabel(nodeSpec *proto.NodeSpec) map[string]*string {
	labels := map[string]*string{
		constants.NodeNameLabel:          &nodeSpec.Name,
		constants.InstanceLabel:          &nodeSpec.Instance,
		constants.NodeLabelSelectorLabel: &nodeSpec.Name,
		"type":                           common.StrPtr("nodegroup")}

	return merge(DefaultTags(), labels, aws.StringMap(nodeSpec.Labels))
}

func ScopeTag() string {
	return fmt.Sprintf("nb-%s", config.Get().Env)
}

//DefaultTags labels/tags which is added to all spawner resources
func DefaultTags() map[string]*string {
	scope := ScopeTag()
	return map[string]*string{
		constants.Scope:        &scope,
		constants.CreatorLabel: &constants.SpawnerServiceLabel,
	}
}
