package constants

//we need address of this vars coz AWS SDK requirements
//calling StrPtr() or aws.String() seems like repetetive

var (
	WorkspaceId = "workspaceid"
	AwsLabel    = "aws"
	AzureLabel  = "azure"
	GcpLabel    = "gcp"
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

const ActualCost string = "ActualCost"

const (
	CostUSD     = "CostUSD"
	ServiceName = "ServiceName"
	TagValue    = "TagValue"
	UsageDate   = "UsageDate"
)

//cred type

const (
	CredAws    = "aws"
	CredAzure  = "azure"
	CredGitPat = "git-pat"
)
