package azure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func diskName(size int32) string {
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("vol-%d-%s", size, t)
}

func (a *AzureController) createVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	disksClient, err := getDisksClient(cred)
	if err != nil {
		a.logger.Errorw("failed to get the disk client", "error", err)
		return nil, err
	}
	size := int32(req.GetSize())
	name := diskName(size)
	tags := labels.DefaultTags()

	a.logger.Infow("creating disk", "name", name, "size", req.Size)

	//if snapshotId is provided

	var creationData *compute.CreationData

	if req.Snapshotid != "" {
		creationData = &compute.CreationData{
			CreateOption: compute.DiskCreateOptionCopy,
			SourceURI:    &req.SnapshotUri,
		}
	} else {

		creationData = &compute.CreationData{
			CreateOption: compute.DiskCreateOptionEmpty,
		}
	}

	// Doc : https://docs.microsoft.com/en-us/rest/api/compute/disks/create-or-update
	future, err := disksClient.CreateOrUpdate(
		ctx,
		cred.ResourceGroup,
		name,
		compute.Disk{
			Sku: &compute.DiskSku{
				// Doc : https://docs.microsoft.com/en-us/rest/api/compute/disks/create-or-update#diskstorageaccounttypes
				Name: "StandardSSD_LRS",
				Tier: to.StringPtr("StandardSSD_LRS"),
			},
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

	a.logger.Debugw("waiting on the future response")
	err = future.WaitForCompletionRef(ctx, disksClient.Client)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get the disk create or update future response")
	}

	res, err := future.Result(*disksClient)
	if err != nil {
		return nil, err
	}
	return &proto.CreateVolumeResponse{
		ResourceUri: *res.ID,
		Volumeid:    name,
	}, nil
}

func (a *AzureController) deleteDisk(ctx context.Context, dc *compute.DisksClient, groupName, name string) error {

	// Doc : https://docs.microsoft.com/en-us/rest/api/compute/disks/delete
	future, err := dc.Delete(
		ctx,
		groupName,
		name,
	)
	if err != nil {
		return errors.Wrap(err, "deleteDisk: aks call failed")
	}

	a.logger.Debugw("waiting on the delete disk future response")
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

func (a *AzureController) deleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {

	name := req.Volumeid

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	disksClient, err := getDisksClient(cred)
	if err != nil {
		a.logger.Errorw("failed to get the disk client", "error", err)
		return nil, err
	}

	a.logger.Infow("deleting disk", "name", name)
	err = a.deleteDisk(ctx, disksClient, cred.ResourceGroup, name)
	if err != nil {
		return nil, err
	}
	return &proto.DeleteVolumeResponse{Deleted: true}, nil
}
