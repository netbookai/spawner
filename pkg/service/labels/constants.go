package labels

import (
	"fmt"
)

type Label string

const (
	Spawner                  = "spawner"
	NBRegionWkspNetworkStack = "nb-region-ntwk-stk"
)

const (
	//NameLabel a special label key
	NameLabel              Label = "Name" //Capital N for Aws
	CreatorLabel           Label = "creator"
	Scope                  Label = "scope"
	ProvisionerLabel       Label = "provisioner"
	ClusterNameLabel       Label = "cluster-name"
	NodeNameLabel          Label = "node-name"
	InstanceLabel          Label = "instance"
	NodeLabelSelectorLabel Label = "nodeLabelSelector"
	VpcTagKey              Label = "vpc"
	NBTypeTagkey           Label = "nb-type"
	ResourceType           Label = "type"
	//WorkspaceLabel           = "workspaceid"

	LabelNamespace = "netbook.ai"
)

func (l Label) Key() string {

	if l == NameLabel {
		return string(l)
	}
	return fmt.Sprintf("%s/%s", LabelNamespace, string(l))
}

func (l Label) KeyPtr() *string {
	k := l.Key()
	return &k
}
