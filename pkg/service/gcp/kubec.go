package gcp

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (g *GCPController) getToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return nil, nil
}

func (g *GCPController) getKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {
	return nil, nil
}
