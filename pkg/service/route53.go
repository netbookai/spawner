package service

import (
	"context"

	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
)

func (svc *spawnerService) addRoute53Record(ctx context.Context, dnsName, recordName, regionName string, isAwsResource bool) (string, error) {
	changeId, err := system.AddRoute53Record(ctx, dnsName, recordName, regionName, isAwsResource)
	if err != nil {
		svc.logger.Error(ctx, "failed to add route53 record", "error", err)
		return "", err
	}

	return changeId, nil
}
