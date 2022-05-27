package labels

import "fmt"

type Label string

//NameLabel a special label key
const NameLabel = "Name" //Capital N for Aws

const (
	CreatorLabel             Label = "creator"
	SpawnerLabel             Label = "spawner"
	Scope                    Label = "scope"
	ProvisionerLabel         Label = "provisioner"
	ClusterNameLabel         Label = "cluster-name"
	NodeNameLabel            Label = "node-name"
	InstanceLabel            Label = "instance"
	NodeLabelSelectorLabel   Label = "nodeLabelSelector"
	VpcTagKey                Label = "vpc"
	NBTypeTagkey             Label = "nb-type"
	NBRegionWkspNetworkStack Label = "nb-region-ntwk-stk"
	//WorkspaceLabel           = "workspaceid"

	LabelNamespace = "netbook.ai"
)

func (l Label) Key() string {
	return fmt.Sprintf("%s/%s", LabelNamespace, string(l))
}
