package service

import (
	"context"

	"github.com/libdns/libdns"
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

func (svc *spawnerService) getRoute53Record(ctx context.Context, dnsName, recordName, regionName string, isAwsResource bool) ([]libdns.Record, error) {
	changeId, err := system.GetRoute53Record(ctx, dnsName, recordName, regionName)
	if err != nil {
		svc.logger.Error(ctx, "failed to add route53 record", "error", err)
		return nil, err
	}

	return changeId, nil
}

func (svc *spawnerService) appendRoute53Records(ctx context.Context, regionName string, records []libdns.Record) ([]libdns.Record, error) {
	changeId, err := system.AppendRecords(ctx, regionName, records)
	if err != nil {
		svc.logger.Error(ctx, "failed to add route53 record", "error", err)
		return nil, err
	}

	return changeId, nil
}

func (svc *spawnerService) deleteRoute53Records(ctx context.Context, regionName string, records []libdns.Record) ([]libdns.Record, error) {
	changeId, err := system.DeleteRecords(ctx, regionName, records)
	if err != nil {
		svc.logger.Error(ctx, "failed to add route53 record", "error", err)
		return nil, err
	}

	return changeId, nil
}
