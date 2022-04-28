package common

import (
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
)

type InstanceSizeMap map[string]string

var providerInstanceType map[string]InstanceSizeMap

func azure() InstanceSizeMap {
	m := InstanceSizeMap{

		"s":       "Standard_B1s",
		"m":       "Standard_F8s_v2",
		"l":       "Standard_F32s_v2",
		"xl":      "Standard_F64_v2",
		"m+t4":    "Standard_NC4as_T4_v3",
		"m+k80":   "Standard_NC6",
		"l+k80":   "Standard_NC12",
		"xl+k80":  "Standard_NC24",
		"m+v100":  "Standard_NC6s_v3",
		"l+v100":  "Standard_NC12s_v3",
		"xl+v100": "Standard_NC24s_v3",
	}
	return m
}

func aws() InstanceSizeMap {
	m := InstanceSizeMap{
		"s":       "t2.micro",
		"m":       "m5.2xlarge",
		"l":       "m5.8xlarge",
		"xl":      "m5.16xlarge",
		"m+t4":    "g4dn.xlarge",
		"m+k80":   "p2.xlarge",
		"l+k80":   "p2.8xlarge",
		"xl+k80":  "p2.16xlarge",
		"m+v100":  "p3.xlarge",
		"l+v100":  "p3.8xlarge",
		"xl+v100": "p3.16xlarge",
	}
	return m
}

func gcp() InstanceSizeMap {
	// https://github.com/iterative/terraform-provider-iterative/blob/master/iterative/gcp/provider.go#L415
	m := InstanceSizeMap{
		"s":       "g1-small",
		"m":       "e2-custom-8-32768",
		"l":       "e2-custom-32-131072",
		"xl":      "n2-custom-64-262144",
		"m+t4":    "n1-standard-4",
		"m+k80":   "custom-8-53248",
		"l+k80":   "custom-32-131072",
		"xl+k80":  "custom-64-212992-ext",
		"m+v100":  "custom-8-65536-ext",
		"l+v100":  "custom-32-262144-ext",
		"xl+v100": "custom-64-524288-ext",
	}
	return m
}

func init() {
	providerInstanceType = make(map[string]InstanceSizeMap)
	providerInstanceType[constants.AzureLabel] = azure()
	providerInstanceType[constants.AwsLabel] = aws()
	providerInstanceType[constants.GcpLabel] = gcp()
}

//GetInstance given machine size return the exact instance type for the provider
func GetInstance(provider, machine string) string {

	p, ok := providerInstanceType[provider]
	if !ok {
		return ""
	}
	if t, ok := p[machine]; ok {
		return t
	}
	return ""
}
