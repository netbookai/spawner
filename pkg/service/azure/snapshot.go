package azure

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-12-01/compute"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (a *azureController) createDiskSnapshot(ctx context.Context, sc *compute.SnapshotsClient, groupName, name string, disk *compute.Disk, region string, tags map[string]*string) (string, error) {

	// Doc : https://docs.microsoft.com/en-us/rest/api/compute/snapshots/create-or-update
	future, err := sc.CreateOrUpdate(
		ctx,
		groupName,
		name,
		compute.Snapshot{
			SnapshotProperties: &compute.SnapshotProperties{
				DiskAccessID: disk.DiskAccessID,
				DiskSizeGB:   disk.DiskSizeGB,

				CreationData: &compute.CreationData{
					CreateOption: compute.DiskCreateOptionCopy,
					SourceURI:    disk.ID,
				},
			},

			Location: &region,
			Tags:     tags,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "createSnapshot: aks call failed")
	}

	a.logger.Debug(ctx, "waiting on the future response")
	err = future.WaitForCompletionRef(ctx, sc.Client)
	if err != nil {
		return "", errors.Wrap(err, "cannot get the creeate snapshot response")
	}

	res, err := future.Result(*sc)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		b, err := io.ReadAll(res.Response.Body)
		if err != nil {
			b = []byte("failed to read response body")
		}
		return "", fmt.Errorf("failed to create snapshot: %s", string(b))
	}
	return *res.ID, nil
}

func (a *azureController) createSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {

	name := fmt.Sprintf("%s-snapshot", req.Volumeid)
	region := req.Region
	tags := labels.DefaultTags()

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	groupName := cred.ResourceGroup

	for k, v := range req.Labels {
		v := v
		tags[k] = &v
	}
	dc, err := getDisksClient(cred)
	if err != nil {
		return nil, err
	}

	disk, err := dc.Get(ctx, groupName, req.Volumeid)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get the disk")
	}
	a.logger.Info(ctx, "creating disk snapshot", "name", name, "source", req.Volumeid)

	sc, err := getSnapshotClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get snapshot client")
	}

	uri, err := a.createDiskSnapshot(ctx, sc, groupName, name, &disk, region, tags)
	if err != nil {
		return nil, err
	}

	return &proto.CreateSnapshotResponse{Snapshotid: name, SnapshotUri: uri}, nil
}

func (a *azureController) createSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {

	name := fmt.Sprintf("%s-snapshot", req.Volumeid)
	region := req.Region

	account := req.AccountName

	cred, err := getCredentials(ctx, account)

	if err != nil {
		return nil, err
	}
	tags := labels.DefaultTags()

	for k, v := range req.Labels {
		v := v
		tags[k] = &v
	}

	dc, err := getDisksClient(cred)
	if err != nil {
		return nil, err
	}

	disk, err := dc.Get(ctx, cred.ResourceGroup, req.Volumeid)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get the disk")
	}
	a.logger.Info(ctx, "creating disk snapshot", "name", name, "source", req.Volumeid)

	sc, err := getSnapshotClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get snapshot client")
	}

	uri, err := a.createDiskSnapshot(ctx, sc, cred.ResourceGroup, name, &disk, region, tags)
	if err != nil {
		return nil, err
	}
	a.logger.Info(ctx, "snapshot created, deleting source disk", "source", *disk.Name)
	err = a.deleteDisk(ctx, dc, cred.ResourceGroup, req.Volumeid)
	if err != nil {
		return nil, err
	}

	return &proto.CreateSnapshotAndDeleteResponse{Snapshotid: name, SnapshotUri: uri}, nil
}

func (a *azureController) deleteSnapshotInternal(ctx context.Context, sc *compute.SnapshotsClient, groupName, snapshotId string) error {

	future, err := sc.Delete(ctx, groupName, snapshotId)
	a.logger.Debug(ctx, "waiting on the delete snapshot future response")
	err = future.WaitForCompletionRef(ctx, sc.Client)
	if err != nil {
		return errors.Wrap(err, "cannot get the snapshot delete response")
	}

	res, err := future.Result(*sc)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return fmt.Errorf("failed to delete snapshot: requested snapshot '%s' not found", snapshotId)
	}

	return nil
}

func (a *azureController) deleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		a.logger.Error(ctx, "failed to get the azure credentials", "error", err)
		return nil, errors.Wrap(err, "deleteSnapshot")
	}

	sc, err := getSnapshotClient(cred)
	if err != nil {
		a.logger.Error(ctx, "faied to get the snapshot client", "error", err)
		return nil, errors.Wrap(err, "deleteSnapshot")
	}
	err = a.deleteSnapshotInternal(ctx, sc, cred.ResourceGroup, req.SnapshotId)
	if err != nil {
		a.logger.Error(ctx, "failed to delete snapshot", "error", err, "snapshotid", req.SnapshotId)
		return nil, errors.Wrap(err, "deleteSnapshot")
	}
	a.logger.Info(ctx, "snapshot deleted", "snapshotid", req.SnapshotId)
	return &proto.DeleteSnapshotResponse{}, nil
}

func (a *azureController) copySnapshot(ctx context.Context, req *proto.CopySnapshotRequest) (*proto.CopySnapshotResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		a.logger.Error(ctx, "failed to get the azure credentials", "error", err)
		return nil, errors.Wrap(err, "copySnapshot")
	}

	sc, err := getSnapshotClient(cred)
	if err != nil {
		a.logger.Error(ctx, "faied to get the snapshot client", "error", err)
		return nil, errors.Wrap(err, "copySnapshot")
	}
	name := fmt.Sprintf("copy-%s", req.SnapshotId)
	region := req.Region
	uri := req.SnapshotUri
	tags := labels.DefaultTags()

	for k, v := range req.Labels {
		v := v
		tags[k] = &v
	}

	res, err := sc.CreateOrUpdate(ctx, cred.ResourceGroup, name, compute.Snapshot{
		SnapshotProperties: &compute.SnapshotProperties{
			CreationData: &compute.CreationData{
				CreateOption:     compute.DiskCreateOptionCopy,
				SourceResourceID: &uri,
			},
		},
		Location: &region,
		Tags:     tags,
	})
	if err != nil {
		a.logger.Error(ctx, "failed to copy snapshot", "error", err)
		return nil, errors.Wrap(err, "createOrUpdate of snapshot failed")
	}
	err = res.WaitForCompletionRef(ctx, sc.Client)
	if err != nil {
		a.logger.Error(ctx, "failed to wait on copy snapshot", "error", err)
		return nil, errors.Wrap(err, "wait on the operation failed")
	}

	snapshot, err := res.Result(*sc)
	if err != nil {

		a.logger.Error(ctx, "failed to get result of copy snapshot", "error", err)
		return nil, errors.Wrap(err, "failed to get the result of operation")
	}
	return &proto.CopySnapshotResponse{
		NewSnapshotId:  name,
		NewSnapshotUri: *snapshot.ID,
	}, nil
}
