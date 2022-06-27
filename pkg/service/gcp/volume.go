package gcp

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	disk_proto "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

func diskName(size int32) string {
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("vol-%d-%s", size, t)
}

//diskType return the URL of the disk type
func diskType(projectId, zone, typ string) string {
	return fmt.Sprintf("projects/%s/zones/%s/diskTypes/%s", projectId, zone, typ)
}

func (g *gcpController) createVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createVolume ")
	}

	client, err := getDiskClient(ctx, cred)
	if err != nil {
		g.logger.Error(ctx, "failed to get disk client", "error", err)
		return nil, errors.Wrap(err, "failed to get disk client")
	}
	//FIXME: hard coded zone info
	zone := fmt.Sprintf("%s-a", req.Region)
	diskType := diskType(cred.ProjectId, zone, "pd-balanced")
	size := req.Size
	name := diskName(int32(size))
	label := make(map[string]string)
	for k, v := range labels.DefaultTags() {
		label[k] = *v
	}

	for k, v := range req.Labels {
		label[k] = v

	}
	idr := disk_proto.InsertDiskRequest{
		DiskResource: &disk_proto.Disk{
			Labels: label,
			Name:   &name,
			SizeGb: &size,
			Type:   &diskType,
		},
		Project: cred.ProjectId,
		Zone:    zone,
	}

	if req.Snapshotid != "" {

		g.logger.Info(ctx, "creating disk from snapshot", "snapshot", req.Snapshotid)
		source := fmt.Sprintf("global/snapshots/%s", req.Snapshotid)
		idr.DiskResource.SourceSnapshot = &source
	}
	g.logger.Info(ctx, "creating volume", "name", name, "size", size)
	// Doc : https://cloud.google.com/compute/docs/reference/rest/v1/disks/insert
	r, err := client.Insert(ctx, &idr)
	if err != nil {
		g.logger.Error(ctx, "failed to insert/create new disk", "error", err)
		return nil, errors.Wrap(err, "disk.Insert failed")

	}
	err = r.Wait(ctx)
	if err != nil {
		g.logger.Error(ctx, "wait on volume create failed")
		return nil, errors.Wrap(err, "create volume wait failed")
	}

	g.logger.Info(ctx, "disk created", "name", name, "zone", zone)

	if req.DeleteSnapshot {
		go func() {

			deleteSnapshotTimeout := time.Duration(10) * time.Second
			deleteCtx, cancel := context.WithTimeout(context.Background(), deleteSnapshotTimeout)
			defer cancel()

			g.logger.Info(deleteCtx, "deleting snapshot", "ID", req.Snapshotid)
			err := g.deleteSnapshotInternal(deleteCtx, cred, req.Snapshotid)

			if err != nil {
				//we will silently log error and return here for now, we dont want to tell the user that volume creation failed in this case.
				g.logger.Error(deleteCtx, "failed to delete the snapshot", "error", err)
				return
			}
			g.logger.Info(deleteCtx, "snapshot deleted", "ID", req.Snapshotid)
		}()
	}
	return &proto.CreateVolumeResponse{
		Volumeid: name,
	}, nil
}

func (g *gcpController) deleteVolumeInternal(ctx context.Context, cred *system.GCPCredential, disk, zone string) error {

	client, err := getDiskClient(ctx, cred)
	if err != nil {
		g.logger.Error(ctx, "failed to get disk client", "error", err)
		return errors.Wrap(err, "failed to get disk client")
	}
	r, err := client.Delete(ctx, &disk_proto.DeleteDiskRequest{
		Disk:    disk,
		Project: cred.ProjectId,
		Zone:    zone,
	})

	if err != nil {
		g.logger.Error(ctx, "failed to delete volume", "error", err)
		return errors.Wrap(err, "failed to delete volume")
	}
	g.logger.Debug(ctx, "wait for volume to be deleted")
	err = r.Wait(ctx)
	if err != nil {
		g.logger.Error(ctx, "wait on volume delete failed")
		return errors.Wrap(err, "delete volume wait failed")
	}
	return nil
}

func (g *gcpController) deleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
	}

	zone := fmt.Sprintf("%s-a", req.Region)
	err = g.deleteVolumeInternal(ctx, cred, req.Volumeid, zone)

	if err != nil {
		return nil, err
	}
	g.logger.Info(ctx, "volume deleted", "volume", req.Volumeid)

	return &proto.DeleteVolumeResponse{}, nil
}
