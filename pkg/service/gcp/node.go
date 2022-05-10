package gcp

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (g *GCPController) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	return nil, nil
}

func (g *GCPController) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return nil, nil
}

func (g *GCPController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	return nil, nil
}
