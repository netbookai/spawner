package azure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/helper"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func diskName(size int32) string {
	return helper.DisplayName(helper.VolumeKind, int64(size))
}

func getDiskSku(vt string) (*compute.DiskSku, error) {
	ds := &compute.DiskSku{}
	// Doc : https://docs.microsoft.com/en-us/rest/api/compute/disks/create-or-update#diskstorageaccounttypes
	switch vt {
	case "Premium_LRS", "Premium_ZRS", "StandardSSD_LRS", "StandardSSD_ZRS", "Standard_LRS", "UltraSSD_LRS":
		ds.Name = compute.DiskStorageAccountTypes(vt)
		ds.Tier = &vt
	default:
		return nil, errors.Errorf("invalid volume type '%s'", vt)
	}
	return ds, nil
}

func (a *azureController) createVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	disksClient, err := getDisksClient(cred)
	if err != nil {
		a.logger.Error(ctx, "failed to get the disk client", "error", err)
		return nil, err
	}
	size := int32(req.GetSize())
	name := diskName(size)
	tags := labels.DefaultTags()

	for k, v := range req.Labels {
		v := v
		tags[k] = &v
	}

	a.logger.Info(ctx, "creating disk", "name", name, "size", req.Size)

	//if snapshotId is provided

	var creationData *compute.CreationData

	if req.SnapshotUri != "" {
		creationData = &compute.CreationData{
			CreateOption: compute.DiskCreateOptionCopy,
			SourceURI:    &req.SnapshotUri,
		}
	} else {

		creationData = &compute.CreationData{
			CreateOption: compute.DiskCreateOptionEmpty,
		}
	}

	sku, err := getDiskSku(req.Volumetype)
	if err != nil {
		return nil, errors.Wrap(err, "createVolume: failed get disk SKU")
	}

	// Doc : https://docs.microsoft.com/en-us/rest/api/compute/disks/create-or-update
	future, err := disksClient.CreateOrUpdate(
		ctx,
		cred.ResourceGroup,
		name,
		compute.Disk{
			Sku:      sku,
			Location: to.StringPtr(req.Region),
			DiskProperties: &compute.DiskProperties{
				CreationData: creationData,
				DiskSizeGB:   &size,
			},
			Tags: tags,
		})
	if err != nil {
		return nil, errors.Wrap(err, "createDisk: aks call failed")
	}

	a.logger.Debug(ctx, "waiting on the future response")
	err = future.WaitForCompletionRef(ctx, disksClient.Client)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get the disk create or update future response")
	}

	res, err := future.Result(*disksClient)
	if err != nil {
		return nil, err
	}

	ret := &proto.CreateVolumeResponse{
		ResourceUri: *res.ID,
		Volumeid:    name,
	}

	if req.DeleteSnapshot {
		//spawn a routine and let it delete
		go func() {
			azureDeleteSnapshotTimeout := time.Duration(10)
			ctx, cancel := context.WithTimeout(context.Background(), azureDeleteSnapshotTimeout)
			defer cancel()

			a.logger.Info(ctx, "deleting snapshot", "ID", req.Snapshotid)
			sc, err := getSnapshotClient(cred)
			if err != nil {
				a.logger.Error(ctx, "failed to get the snapshot client", "error", err)
				return
			}

			err = a.deleteSnapshotInternal(ctx, sc, cred.ResourceGroup, req.Snapshotid)
			if err != nil {
				//we will silently log error and return here for now, we dont want to tell the user that volume creation failed in this case.
				a.logger.Error(ctx, "failed to delete the snapshot", "error", err)
				return
			}
			a.logger.Info(ctx, "snapshot deleted", "ID", req.Snapshotid)
		}()
	}
	return ret, nil

}

func (a *azureController) deleteDisk(ctx context.Context, dc *compute.DisksClient, groupName, name string) error {

	// Doc : https://docs.microsoft.com/en-us/rest/api/compute/disks/delete
	future, err := dc.Delete(
		ctx,
		groupName,
		name,
	)
	if err != nil {
		return errors.Wrap(err, "deleteDisk: aks call failed")
	}

	a.logger.Debug(ctx, "waiting on the delete disk future response")
	err = future.WaitForCompletionRef(ctx, dc.Client)
	if err != nil {
		return errors.Wrap(err, "cannot get the disk delete response")
	}

	res, err := future.Result(*dc)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return fmt.Errorf("failed to delete volume: requested volume '%s' not found", name)
	}
	return nil
}

func (a *azureController) deleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {

	name := req.Volumeid

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	disksClient, err := getDisksClient(cred)
	if err != nil {
		a.logger.Error(ctx, "failed to get the disk client", "error", err)
		return nil, err
	}

	a.logger.Info(ctx, "deleting disk", "name", name)
	err = a.deleteDisk(ctx, disksClient, cred.ResourceGroup, name)
	if err != nil {
		return nil, err
	}
	return &proto.DeleteVolumeResponse{Deleted: true}, nil
}
