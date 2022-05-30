package aws

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (a *AWSController) GetElasticRegistryAuth(ctx context.Context, in *proto.GetElasticRegistryAuthRequest) (*proto.GetElasticRegistryAuthResponse, error) {
	return &proto.GetElasticRegistryAuthResponse{}, nil
}
