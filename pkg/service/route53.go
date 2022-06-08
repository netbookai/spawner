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

func (svc *spawnerService) getRoute53TXTRecords(ctx context.Context, regionName string) ([]types.Route53Record, error) {
	records, err := system.GetRoute53TXTRecords(ctx, regionName)
	if err != nil {
		svc.logger.Error(ctx, "failed to get route53 record", "error", err)
		return nil, err
	}

	return records, nil
}

func (svc *spawnerService) appendRoute53Records(ctx context.Context, regionName string, records []types.Route53Record) ([]types.Route53Record, error) {
	appendedRecords, err := system.AppendRoute53Records(ctx, regionName, records)
	if err != nil {
		svc.logger.Error(ctx, "failed to append route53 record", "error", err)
		return nil, err
	}

	return appendedRecords, nil
}

func (svc *spawnerService) deleteRoute53Records(ctx context.Context, regionName string, records []types.Route53Record) ([]types.Route53Record, error) {
	deletedRecords, err := system.DeleteRoute53Records(ctx, regionName, records)
	if err != nil {
		svc.logger.Error(ctx, "failed to delete route53 record", "error", err)
		return nil, err
	}

	return deletedRecords, nil
}
