package azure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func diskName(region string, size int32) string {
	t := time.Now().Format("20060102150405")

	return fmt.Sprintf("vol-%s-%d-%s", region, size, t)
}

func (a *AzureController) createVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {

	groupName := config.Get().AzureResourceGroup

	disksClient, err := getDisksClient()
	if err != nil {
		a.logger.Errorw("failed to get the disk client", "error", err)
		return nil, err
	}
	name := diskName(req.Region, int32(req.GetSize()))
	tags := labels.DefaultTags()

	a.logger.Infow("creating disk", "name", name, "size", req.Size)
	future, err := disksClient.CreateOrUpdate(
		ctx,
		groupName,
		name,
		compute.Disk{
			Location: to.StringPtr(req.Region),
			DiskProperties: &compute.DiskProperties{

				CreationData: &compute.CreationData{
					CreateOption: compute.DiskCreateOptionEmpty,
				},
				DiskSizeGB: to.Int32Ptr(64),
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
	fmt.Println("CreateOrUpdate future ", *res.ID)
	spew.Dump(res)
	return &proto.CreateVolumeResponse{
		Volumeid: name,
	}, nil
}

func (a *AzureController) deleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	groupName := config.Get().AzureResourceGroup

	disksClient, err := getDisksClient()
	if err != nil {
		a.logger.Errorw("failed to get the disk client", "error", err)
		return nil, err
	}

	name := req.Volumeid
	a.logger.Infow("deleting disk", "name", name)
	future, err := disksClient.Delete(
		ctx,
		groupName,
		name,
	)
	if err != nil {
		return nil, errors.Wrap(err, "deleteDisk: aks call failed")
	}

	a.logger.Debugw("waiting on the future response")
	err = future.WaitForCompletionRef(ctx, disksClient.Client)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get the disk delete response")
	}

	res, err := future.Result(*disksClient)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("failed to delete volume: requested volume '%s' not found", name)
	}
	return &proto.DeleteVolumeResponse{
		Deleted: true,
	}, nil
}
