package labels

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
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
		NodeNameLabel.Key():          &nodeSpec.Name,
		InstanceLabel.Key():          &instance,
		NodeLabelSelectorLabel.Key(): &nodeSpec.Name,
		ResourceType.Key():           aws.String("nodegroup")}

	return merge(DefaultTags(), labels, aws.StringMap(nodeSpec.Labels))
}

func ScopeTag() string {
	return fmt.Sprintf("nb-%s", config.Get().Env)
}

//DefaultTags labels/tags which is added to all spawner resources
func DefaultTags() map[string]*string {
	scope := ScopeTag()
	return map[string]*string{
		Scope.Key():        &scope,
		CreatorLabel.Key(): aws.String(Spawner),
	}
}
