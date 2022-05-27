package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/netbookai/log"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func logError(ctx context.Context, methodName string, logger log.Logger, err error) {

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logger.Error(ctx, "Error in ", "method : ", methodName, "error", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Error(ctx, "Error in ", "method : ", methodName, "error", err.Error())
		}
		logger.Error(ctx, "Error in ", "method : ", methodName, "error", err)
	}
}

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
func (svc AWSController) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
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

	//creating session
	session, err := NewSession(ctx, region, req.AccountName)

	if err != nil {
		logger.Error(ctx, "Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()
	//calling aws sdk CreateVolume function
	result, err := ec2Client.CreateVolume(input)
	if err != nil {
		logError(ctx, "CreateVolume", logger, err)
		return &proto.CreateVolumeResponse{}, err
	}

	err = ec2Client.WaitUntilVolumeAvailable(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{result.VolumeId},
	})
	if err != nil {
		logError(ctx, "WaitForVolumeAvailable", logger, err)
		return &proto.CreateVolumeResponse{}, err
	}

	res := &proto.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
		Error:    "",
	}

	//if delete requested,nuke em
	if req.DeleteSnapshot {
		go func() {

			//this is to handle the aws API call timeout, we wont need to handle the routine timeout here
			awsDeleteSnapshotTimeout := time.Duration(10)
			ctx, cancel := context.WithTimeout(context.Background(), awsDeleteSnapshotTimeout)
			defer cancel()

			svc.logger.Info(ctx, "deleting snapshot", "ID", snapshotId)
			_, err = ec2Client.DeleteSnapshotWithContext(ctx, &ec2.DeleteSnapshotInput{
				SnapshotId: &snapshotId,
			})

			if err != nil {
				//we will silently log error and return here for now, we dont want to tell the user that volume creation failed in this case.
				svc.logger.Error(ctx, "failed to delete the snapshot", "error", err)
			}
			svc.logger.Info(ctx, "snapshot deleted", "ID", snapshotId)
		}()
	}

	return res, nil
}

//DeleteVolume delete aws volume
func (svc AWSController) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
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
	_, err = ec2Client.DeleteVolume(input)

	if err != nil {
		logError(ctx, "DeleteVolume", logger, err)
		return &proto.DeleteVolumeResponse{}, err
	}

	//note: since now err is nil so assigning deleted = true
	res := &proto.DeleteVolumeResponse{
		Deleted: true,
	}

	return res, nil
}

//CreateSnapshot create volume snapshot
func (svc AWSController) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
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
	result, err := ec2Client.CreateSnapshot(input)
	if err != nil {
		logError(ctx, "CreateSnapshot", logger, err)
		return &proto.CreateSnapshotResponse{}, err
	}

	err = ec2Client.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{result.SnapshotId},
	})
	if err != nil {
		logError(ctx, "WaitForSnapshotCompleted", logger, err)
		return &proto.CreateSnapshotResponse{}, err
	}

	res := &proto.CreateSnapshotResponse{
		Snapshotid: *result.SnapshotId,
	}

	return res, nil
}

//CreateSnapshotAndDelete create a  snapshot of volume and delete the volume
func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
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
	resultSnapshot, err := ec2Client.CreateSnapshot(inputSnapshot)
	if err != nil {
		logError(ctx, "CreateSnapshot", logger, err)
		return &proto.CreateSnapshotAndDeleteResponse{}, err
	}

	err = ec2Client.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{resultSnapshot.SnapshotId},
	})
	if err != nil {
		logError(ctx, "WaitForSnapshotCompleted", logger, err)
		return &proto.CreateSnapshotAndDeleteResponse{}, err
	}

	//inputs for deleteing volume
	inputDelete := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err = ec2Client.DeleteVolume(inputDelete)

	if err != nil {
		logError(ctx, "DeleteVolume", logger, err)

		return &proto.CreateSnapshotAndDeleteResponse{
			Snapshotid: *resultSnapshot.SnapshotId,
			Deleted:    false,
		}, err
	}

	//note: since now err is nil so assigning deleted = true
	res := &proto.CreateSnapshotAndDeleteResponse{
		Snapshotid: *resultSnapshot.SnapshotId,
		Deleted:    true,
	}

	return res, nil
}
