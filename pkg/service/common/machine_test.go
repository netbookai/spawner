package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetInstance(t *testing.T) {

	got := GetInstance("aws", "m")
	assert.Equal(t, "m5.2xlarge", got, "GetInstance('aws', 'm')")

	got = GetInstance("gcp", "xl+k80")
	assert.Equal(t, "custom-64-212992-ext", got, "GetInstance('gcp', 'xl+k80')")

	got = GetInstance("azure", "m+v100")
	assert.Equal(t, "Standard_NC6s_v3", got, "GetInstance('azure', 'm+v100')")

	got = GetInstance("k8s", "m+v100")
	assert.Equal(t, "", got, "GetInstance: for invalid provider")

	got = GetInstance("gcp", "m+v1000")
	assert.Equal(t, "", got, "GetInstance: for invalid machine size")

	assert.True(t, IsGPU(Lk80), "expected gpu machine")
	assert.False(t, IsGPU(M), "expected non-gpu machine")
}
