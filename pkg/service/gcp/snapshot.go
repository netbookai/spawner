package gcp

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	disk_proto "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

func (g *GCPController) createSnapshotInternal(ctx context.Context, cred *system.GCPCredential, disk, zone string, reqLabels map[string]string) (string, error) {

	client, err := getDiskClient(ctx, cred)
	if err != nil {
		g.logger.Error(ctx, "failed to get disk client", "error", err)
		return "", errors.Wrap(err, "failed to get disk client")
	}
	snapshotName := fmt.Sprintf("%s-snapshot", disk)

	label := make(map[string]string)
	for k, v := range labels.DefaultTags() {
		label[k] = *v
	}

	for k, v := range reqLabels {
		label[k] = v
	}
	crr := disk_proto.CreateSnapshotDiskRequest{
		Disk:    disk,
		Project: cred.ProjectId,
		SnapshotResource: &disk_proto.Snapshot{
			Labels:     label,
			Name:       &snapshotName,
			SourceDisk: &disk,
		},
		Zone: zone,
	}
	// Doc : https://cloud.google.com/compute/docs/reference/rest/v1/disks/createSnapshot
	r, err := client.CreateSnapshot(ctx, &crr)
	if err != nil {
		g.logger.Error(ctx, "failed to create a snapshot", "error", err, "name", snapshotName, "disk", disk)
		return "", errors.Wrap(err, "createSnapshot")
	}
	g.logger.Info(ctx, "waiting for snapshot creation complete", "name", snapshotName)
	err = r.Wait(ctx)
	if err != nil {
		g.logger.Error(ctx, "failed to wait till snapshot creation complete", "error", err)
		return "", errors.Wrap(err, "createSnapshot wait")
	}
	return snapshotName, nil
}

func (g *GCPController) createSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createSnapshot ")
	}
	//FIXME: hard coded zone info
	zone := fmt.Sprintf("%s-a", req.Region)

	name, err := g.createSnapshotInternal(ctx, cred, req.Volumeid, zone, req.Labels)
	if err != nil {
		return nil, err
	}
	g.logger.Info(ctx, "snapshot created", "name", name, "zone", zone)
	return &proto.CreateSnapshotResponse{
		Snapshotid: name,
	}, nil
}

func (g *GCPController) createSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createSnapshot ")
	}
	//FIXME: hard coded zone info
	zone := fmt.Sprintf("%s-a", req.Region)

	name, err := g.createSnapshotInternal(ctx, cred, req.Volumeid, zone, req.Labels)
	if err != nil {
		return nil, errors.Wrap(err, "CreateSnapshotAndDelete")
	}
	g.logger.Info(ctx, "snapshot created", "name", name, "zone", zone)

	err = g.deleteVolumeInternal(ctx, cred, req.Volumeid, zone)
	if err != nil {
		return &proto.CreateSnapshotAndDeleteResponse{
			Snapshotid: name,
			Deleted:    false,
		}, err
	}
	g.logger.Info(ctx, "volume deleted", "volume", req.Volumeid, "zone", zone)
	return &proto.CreateSnapshotAndDeleteResponse{
		Snapshotid: name,
		Deleted:    true,
	}, nil
}

func (g *GCPController) deleteSnapshotInternal(ctx context.Context, cred *system.GCPCredential, snapshot string) error {

	client, err := getSnapshotClient(ctx, cred)
	if err != nil {
		g.logger.Error(ctx, "failed to get snapshot client", "error", err)
		return errors.Wrap(err, "failed to get snapshot client")
	}
	crr := disk_proto.DeleteSnapshotRequest{
		Project:  cred.ProjectId,
		Snapshot: snapshot,
	}
	// Doc : https://cloud.google.com/compute/docs/reference/rest/v1/snapshots/delete
	r, err := client.Delete(ctx, &crr)
	if err != nil {
		g.logger.Error(ctx, "failed to delete a snapshot", "error", err, "name", snapshot)
		return errors.Wrap(err, "deleteSnapshot")
	}
	g.logger.Info(ctx, "waiting for snapshot delete complete", "name", snapshot)
	err = r.Wait(ctx)
	if err != nil {
		g.logger.Error(ctx, "failed to wait till snapshot deletion", "error", err)
		return errors.Wrap(err, "deleteSnapshot wait")
	}
	return nil
}

func (g *GCPController) deleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
	}
	err = g.deleteSnapshotInternal(ctx, cred, req.SnapshotId)
	if err != nil {
		g.logger.Info(ctx, "snapshot deleted", "snapshot", req.SnapshotId)
	}
	return &proto.DeleteSnapshotResponse{}, err
}
