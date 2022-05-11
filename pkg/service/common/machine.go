package common

import (
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
)

//machine type constants

const S = "s"
const M = "m"
const L = "l"
const XL = "xl"
const MT4 = "m+t4"
const Mk80 = "m+k80"
const Lk80 = "l+k80"
const XLk80 = "xl+k80"
const Mv100 = "m+v100"
const Lv100 = "l+v100"
const XLv100 = "xl+v100"

type InstanceSizeMap map[string]string

var providerInstanceType map[string]InstanceSizeMap
var gpus map[string]bool

func azure() InstanceSizeMap {
	m := InstanceSizeMap{

		S:      "Standard_B1s",
		M:      "Standard_F8s_v2",
		L:      "Standard_F32s_v2",
		XL:     "Standard_F64_v2",
		MT4:    "Standard_NC4as_T4_v3",
		Mk80:   "Standard_NC6",
		Lk80:   "Standard_NC12",
		XLk80:  "Standard_NC24",
		Mv100:  "Standard_NC6s_v3",
		Lv100:  "Standard_NC12s_v3",
		XLv100: "Standard_NC24s_v3",
	}
	return m
}

func aws() InstanceSizeMap {
	m := InstanceSizeMap{
		S:      "t2.micro",
		M:      "m5.2xlarge",
		L:      "m5.8xlarge",
		XL:     "m5.16xlarge",
		MT4:    "g4dn.xlarge",
		Mk80:   "p2.xlarge",
		Lk80:   "p2.8xlarge",
		XLk80:  "p2.16xlarge",
		Mv100:  "p3.xlarge",
		Lv100:  "p3.8xlarge",
		XLv100: "p3.16xlarge",
	}
	return m
}

func gcp() InstanceSizeMap {
	// https://github.com/iterative/terraform-provider-iterative/blob/master/iterative/gcp/provider.go#L415
	m := InstanceSizeMap{
		S:      "g1-small",
		M:      "e2-custom-8-32768",
		L:      "e2-custom-32-131072",
		XL:     "n2-custom-64-262144",
		MT4:    "n1-standard-4",
		Mk80:   "custom-8-53248",
		Lk80:   "custom-32-131072",
		XLk80:  "custom-64-212992-ext",
		Mv100:  "custom-8-65536-ext",
		Lv100:  "custom-32-262144-ext",
		XLv100: "custom-64-524288-ext",
	}
	return m
}

func init() {
	providerInstanceType = make(map[string]InstanceSizeMap)
	providerInstanceType[constants.AzureLabel] = azure()
	providerInstanceType[constants.AwsLabel] = aws()
	providerInstanceType[constants.GcpLabel] = gcp()

	gpus = map[string]bool{
		MT4:    true,
		Mk80:   true,
		Lk80:   true,
		XLk80:  true,
		Mv100:  true,
		Lv100:  true,
		XLv100: true,
	}
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

func IsGPU(m string) bool {
	return gpus[m]
}
