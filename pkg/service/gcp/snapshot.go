package gcp

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	disk_proto "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

func (g *GCPController) createSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	// Doc : https://cloud.google.com/compute/docs/reference/rest/v1/disks/createSnapshot

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
	}

	client, err := getDiskClient(ctx, cred)
	if err != nil {
		g.logger.Error(ctx, "failed to get disk client", "error", err)
		return nil, errors.Wrap(err, "failed to get disk client")
	}
	disk := req.Volumeid
	snapshotName := fmt.Sprintf("%s-snapshot", req.Volumeid)
	crr := disk_proto.CreateSnapshotDiskRequest{
		Disk:    disk,
		Project: cred.ProjectId,
		SnapshotResource: &disk_proto.Snapshot{
			Labels:     map[string]string{},
			Name:       &snapshotName,
			SourceDisk: &disk,
		},
		Zone: req.Region,
	}
	r, err := client.CreateSnapshot(ctx, &crr)
	if err != nil {
		g.logger.Error(ctx, "failed to create a snapshot", "error", err, "name", snapshotName, "disk", disk)
		return nil, errors.Wrap(err, "createSnapshot")
	}
	g.logger.Info(ctx, "waiting for snapshot creation complete", "name", snapshotName)
	err = r.Wait(ctx)
	if err != nil {
		g.logger.Error(ctx, "failed to wait till snapshot creation complete", "error", err)
		return nil, errors.Wrap(err, "createSnapshot wait")
	}
	g.logger.Info(ctx, "snapshot created", "name", snapshotName)
	return &proto.CreateSnapshotResponse{}, nil
}

func (g *GCPController) createSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return nil, nil
}

func (g *GCPController) deleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {
	return nil, nil
}
