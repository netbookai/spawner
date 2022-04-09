package constants

//we need address of this vars coz AWS SDK requirements
//calling StrPtr() or aws.String() seems like repetetive

var (
	NameLabel           string = "Name" //Capital N for Aws
	CreatorLabel               = "creator"
	SpawnerServiceLabel        = "spawner-service"
	Scope                      = "scope"

	ProvisionerLabel         = "provisioner"
	ClusterNameLabel         = "cluster-name"
	WorkspaceLabel           = "workspaceid"
	NodeNameLabel            = "node-name"
	InstanceLabel            = "instance"
	NodeLabelSelectorLabel   = "nodeLabelSelector"
	AwsLabel                 = "aws"
	VpcTagKey                = "vpc"
	NBTypeTagkey             = "nb-type"
	NBRegionWkspNetworkStack = "nb-region-ntwk-stk"
	WorkspaceId              = "workspaceid"
	KubeVersion              = "1.20"
	AzureKubeVersion         = "1.20.15"
	AzureLabel               = "azure"
)

type CloudProvider string

const (
	AwsCloud   CloudProvider = "aws"
	AzureCloud CloudProvider = "azure"
	GcpCloud   CloudProvider = "gcp"
)

const (
	Active   = "active"
	Inactive = "inactive"
)
