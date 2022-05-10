package gcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"go.uber.org/zap"
)

func Test_createCluster(t *testing.T) {
	gcp := NewController(zap.L().Sugar())

	req := &proto.ClusterRequest{
		Provider:    "gcp",
		Region:      "us-central1",
		AccountName: "",
		ClusterName: "gcp-cluster-test-1",
		Node: &proto.NodeSpec{
			Name:       "nodepool-gcp",
			Instance:   "e2-medium",
			DiskSize:   30,
			Labels:     map[string]string{"creator_proxy": "test"},
			GpuEnabled: false,
			Count:      1,
		},
		Labels: map[string]string{},
	}
	resp, err := gcp.CreateCluster(context.Background(), req)

	assert.Nil(t, err, "create Cluster failed")
	assert.NotNil(t, resp, "create Cluster response is nil")
}
