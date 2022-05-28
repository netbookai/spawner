package gcp

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	disk_proto "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

func diskName(size int32) string {
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("vol-%d-%s", size, t)
}

func diskType(projectId, zone, typ string) string {
	return fmt.Sprintf("projects/%s/zones/%s/diskTypes/%s", projectId, zone, typ)
}

func (g *GCPController) createVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
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
			//SourceSnapshot:              new(string),
		},
		Project: cred.ProjectId,
		//create from snapshot ?
		//SourceImage:  new(string),
		Zone: zone,
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
	return &proto.CreateVolumeResponse{
		Volumeid: name,
	}, nil
}

func (g *GCPController) deleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "createCluster ")
	}

	client, err := getDiskClient(ctx, cred)
	if err != nil {
		g.logger.Error(ctx, "failed to get disk client", "error", err)
		return nil, errors.Wrap(err, "failed to get disk client")
	}
	zone := fmt.Sprintf("%s-a", req.Region)

	g.logger.Info(ctx, "deleting volume", "volume", req.Volumeid)
	r, err := client.Delete(ctx, &disk_proto.DeleteDiskRequest{
		Disk:    req.Volumeid,
		Project: cred.ProjectId,
		Zone:    zone,
	})

	if err != nil {
		g.logger.Error(ctx, "failed to delete volume", "error", err)
		return nil, errors.Wrap(err, "failed to delete volume")
	}
	g.logger.Debug(ctx, "wait for volume to be deleted")
	err = r.Wait(ctx)
	if err != nil {
		g.logger.Error(ctx, "wait on volume delete failed")
		return nil, errors.Wrap(err, "delete volume wait failed")
	}

	return &proto.DeleteVolumeResponse{}, nil
}
