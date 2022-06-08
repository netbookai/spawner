package service

import (
	"context"

	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	"gitlab.com/netbook-devs/spawner-service/pkg/types"
)

func (svc *spawnerService) addRoute53Record(ctx context.Context, dnsName, recordName, regionName string, isAwsResource bool) (string, error) {
	changeId, err := system.AddRoute53Record(ctx, dnsName, recordName, regionName, isAwsResource)
	if err != nil {
		svc.logger.Error(ctx, "failed to add route53 record", "error", err)
		return "", err
	}

	return changeId, nil
}

func (svc *spawnerService) getRoute53TXTRecords(ctx context.Context) ([]types.Route53Record, error) {
	records, err := system.GetRoute53TXTRecords(ctx)
	if err != nil {
		svc.logger.Error(ctx, "failed to get route53 record", "error", err)
		return nil, err
	}

	return records, nil
}

func (svc *spawnerService) appendRoute53Records(ctx context.Context, records []types.Route53Record) error {
	err := system.AppendRoute53Records(ctx, records)
	if err != nil {
		svc.logger.Error(ctx, "failed to append route53 record", "error", err)
		return err
	}

	return nil
}

func (svc *spawnerService) deleteRoute53Records(ctx context.Context, records []types.Route53Record) error {
	err := system.DeleteRoute53Records(ctx, records)
	if err != nil {
		svc.logger.Error(ctx, "failed to delete route53 record", "error", err)
		return err
	}

	return nil
}
