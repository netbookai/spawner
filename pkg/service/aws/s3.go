package aws

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (aws *awsController) PresignS3Url(ctx context.Context, in *proto.PresignS3UrlRequest) (*proto.PresignS3UrlResponse, error) {
	return &proto.PresignS3UrlResponse{}, nil
}
