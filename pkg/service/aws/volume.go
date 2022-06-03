package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func awsTags(label map[string]string) []*ec2.Tag {
	for k, v := range labels.DefaultTags() {
		label[k] = *v
	}

	tags := []*ec2.Tag{}
	for key, value := range label {
		tags = append(tags, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tags
}

//CreateVolume create aws volume
func (svc awsController) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	//Creates an EBS volume

	logger := svc.logger

	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()
	region := req.Region
	labels := req.GetLabels()

	if labels == nil {
		labels = map[string]string{}
	}

	//creating session
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		logger.Error(ctx, "failed to create a new aws session", "error", err)
		return nil, errors.Wrap(err, "CreateVolume ")
	}

	ec2Client := session.getEC2Client()

	tags := awsTags(labels)
	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availabilityZone),
		VolumeType:       aws.String(volumeType),
		Size:             aws.Int64(size),
		SnapshotId:       aws.String(snapshotId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeVolume),
				Tags:         tags,
			},
		},
	}
	//calling aws sdk CreateVolume function
	result, err := ec2Client.CreateVolumeWithContext(ctx, input)
	if err != nil {
		logger.Error(ctx, "failed to create volume", "error", err)
		return nil, errors.Wrap(err, "CreateVolume ")
	}

	err = ec2Client.WaitUntilVolumeAvailableWithContext(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: []*string{result.VolumeId},
	})

	if err != nil {
		logger.Error(ctx, "failed to wait till volume is available", "error", err)
		return nil, errors.Wrap(err, "CreateVolume ")
	}

	res := &proto.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
	}

	//if delete requested,nuke em
	if req.DeleteSnapshot {
		go func() {
			//this is to handle the aws API call timeout, we wont need to handle the routine timeout here
			awsDeleteSnapshotTimeout := time.Duration(10)
			ctx, cancel := context.WithTimeout(context.Background(), awsDeleteSnapshotTimeout)
			defer cancel()

			logger.Info(ctx, "deleting snapshot", "ID", snapshotId)
			_, err = ec2Client.DeleteSnapshotWithContext(ctx, &ec2.DeleteSnapshotInput{
				SnapshotId: &snapshotId,
			})

			if err != nil {
				//we will silently log error and return here for now, we dont want to tell the user that volume creation failed in this case.
				logger.Error(ctx, "failed to delete the snapshot", "error", err)
			}
			logger.Info(ctx, "snapshot deleted", "ID", snapshotId)
		}()
	}

	return res, nil
}

//DeleteVolume delete aws volume
func (svc awsController) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	//Deletes an EBS volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//creating session
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		logger.Error(ctx, "Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()

	if err != nil {
		logger.Error(ctx, "Can't start AWS session", "error", err)
		return nil, err
	}
	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err = ec2Client.DeleteVolumeWithContext(ctx, input)

	if err != nil {
		logger.Error(ctx, "filed to delete volume", "error", err, "id", volumeid)
		return nil, errors.Wrap(err, "DeleteVolume")
	}

	//note: since now err is nil so assigning deleted = true
	res := &proto.DeleteVolumeResponse{
		Deleted: true,
	}

	return res, nil
}

//CreateSnapshot create volume snapshot
func (svc awsController) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	//Creates a Snapshot of a volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	labels := req.GetLabels()

	if labels == nil {
		labels = map[string]string{}
	}

	tags := awsTags(labels)

	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeSnapshot),
				Tags:         tags,
			},
		},
	}

	//creating session
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		logger.Error(ctx, "Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()

	//calling aws sdk method to snapshot volume
	result, err := ec2Client.CreateSnapshotWithContext(ctx, input)
	if err != nil {
		logger.Error(ctx, "failed to create a snapshot", "error", err, "volumeid", volumeid)
		return nil, errors.Wrap(err, "CreateSnapshot")
	}

	logger.Info(ctx, "created snapshot", "snapshot-id", result.SnapshotId)

	err = ec2Client.WaitUntilSnapshotCompletedWithContext(ctx, &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{result.SnapshotId},
	})
	if err != nil {
		logger.Error(ctx, "failed to wait on snapshot completion", "error", err)
		return nil, errors.Wrap(err, "CreateSnapshot")
	}

	res := &proto.CreateSnapshotResponse{
		Snapshotid: *result.SnapshotId,
	}

	return res, nil
}

//CreateSnapshotAndDelete create a  snapshot of volume and delete the volume
func (svc awsController) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	//First Creates the snapshot of the volume then deletes the volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	labels := req.GetLabels()

	if labels == nil {
		labels = map[string]string{}
	}

	tags := awsTags(labels)

	inputSnapshot := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeSnapshot),
				Tags:         tags,
			},
		},
	}

	//creating session
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		logger.Error(ctx, "Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()
	//calling aws sdk CreateSnapshot method
	resultSnapshot, err := ec2Client.CreateSnapshotWithContext(ctx, inputSnapshot)
	if err != nil {
		logger.Error(ctx, "failed to create a snapshot", "error", err, "volumeid", volumeid)
		return nil, errors.Wrap(err, "CreateSnapshotAndDelete")
	}

	err = ec2Client.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{resultSnapshot.SnapshotId},
	})
	if err != nil {
		logger.Error(ctx, "failed to wait on snapshot completion", "error", err, "snapshotid", resultSnapshot.SnapshotId)
		return nil, errors.Wrap(err, "CreateSnapshotAndDelete")
	}

	//inputs for deleteing volume
	inputDelete := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err = ec2Client.DeleteVolumeWithContext(ctx, inputDelete)

	if err != nil {
		logger.Error(ctx, "failed to delete volume", "error", err, "volumeid", volumeid)

		return &proto.CreateSnapshotAndDeleteResponse{
			Snapshotid: *resultSnapshot.SnapshotId,
			Deleted:    false,
		}, errors.Wrap(err, "failed to delete the backing volume")
	}

	//note: since now err is nil so assigning deleted = true
	res := &proto.CreateSnapshotAndDeleteResponse{
		Snapshotid: *resultSnapshot.SnapshotId,
		Deleted:    true,
	}

	return res, nil
}

func (a *awsController) DeleteSnapshot(ctx context.Context, req *proto.DeleteSnapshotRequest) (*proto.DeleteSnapshotResponse, error) {

	session, err := NewSession(ctx, req.Region, req.AccountName)

	if err != nil {
		a.logger.Error(ctx, "Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()
	_, err = ec2Client.DeleteSnapshotWithContext(ctx, &ec2.DeleteSnapshotInput{
		SnapshotId: &req.SnapshotId,
	})

	if err != nil {
		a.logger.Error(ctx, "failed to delete snapshot", "error", err, "snapshotid", req.SnapshotId)
		return nil, errors.Wrap(err, "DeleteSnapshot")
	}
	a.logger.Info(ctx, "snapshot deleted", "snapshotid", req.SnapshotId)

	return &proto.DeleteSnapshotResponse{}, nil
}
