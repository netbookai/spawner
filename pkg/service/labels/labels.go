package labels

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
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

//MergeRequestLabel merges the MergeRequestLabel to tags by taking pointer to the each value of a corresponding key in MergeRequestLabel
//tags will be modified after this call.
func MergeRequestLabel(tags map[string]*string, requestlabels map[string]string) {

	for k, v := range requestlabels {
		//need a copy of the value in v and use that pointer.
		// Using v would result in consuming the latest updated value in v, which is last element in the list
		copyVal := v
		tags[k] = &copyVal
	}

}

func GetNodeLabel(nodeSpec *proto.NodeSpec) map[string]*string {

	instance := ""
	if nodeSpec.MachineType != "" {
		instance = nodeSpec.MachineType
		//+ is not allowed in tag value regex
		instance = strings.Replace(instance, "+", "-", 2)
	}
	if nodeSpec.Instance != "" {
		instance = nodeSpec.Instance
	}

	labels := map[string]*string{
		constants.NodeNameLabel:          &nodeSpec.Name,
		constants.InstanceLabel:          &instance,
		constants.NodeLabelSelectorLabel: &nodeSpec.Name,
		"type":                           aws.String("nodegroup")}

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
