package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/maps"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
	"go.uber.org/zap"
)

func LogError(methodName string, logger *zap.SugaredLogger, err error) {

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logger.Errorw("Error in ", "method : ", methodName, "error", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Errorw("Error in ", "method : ", methodName, "error", err.Error())
		}
		logger.Errorw("Error in ", "method : ", methodName, "error", err)
	}
}

func addAWSTags(labels map[string]string) []*ec2.Tag {

	tagsMap := maps.SimpleReplaceMerge(map[string]string{constants.CREATOR_LABEL: constants.SPAWNER_SERVICE_LABEL, constants.PROVISIONER_LABEL: constants.AWS_LABEL}, labels)
	tags := []*ec2.Tag{}
	for key, value := range tagsMap {
		tags = append(tags, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tags
}

func (svc AWSController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	//Creates an EBS volume

	logger := svc.logger

	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()
	region := req.Region
	tags := addAWSTags(req.GetLabels())

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
	session, err := NewSession(svc.config, region, req.AccountName)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()
	//calling aws sdk CreateVolume function
	result, err := ec2Client.CreateVolume(input)
	if err != nil {
		LogError("CreateVolume", logger, err)
		return &pb.CreateVolumeResponse{}, err
	}

	err = ec2Client.WaitUntilVolumeAvailable(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{result.VolumeId},
	})
	if err != nil {
		LogError("WaitForVolumeAvailable", logger, err)
		return &pb.CreateVolumeResponse{}, err
	}

	res := &pb.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
		Error:    "",
	}

	return res, nil
}

func (svc AWSController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	//Deletes an EBS volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region

	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//creating session
	session, err := NewSession(svc.config, region, req.AccountName)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}
	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err = ec2Client.DeleteVolume(input)

	if err != nil {
		LogError("DeleteVolume", logger, err)
		return &pb.DeleteVolumeResponse{}, err
	}

	//note: since now err is nil so assigning deleted = true
	res := &pb.DeleteVolumeResponse{
		Deleted: true,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	//Creates a Snapshot of a volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region
	tags := addAWSTags(req.GetLabels())

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
	session, err := NewSession(svc.config, region, req.AccountName)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()

	//calling aws sdk method to snapshot volume
	result, err := ec2Client.CreateSnapshot(input)
	if err != nil {
		LogError("CreateSnapshot", logger, err)
		return &pb.CreateSnapshotResponse{}, err
	}

	err = ec2Client.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{result.SnapshotId},
	})
	if err != nil {
		LogError("WaitForSnapshotCompleted", logger, err)
		return &pb.CreateSnapshotResponse{}, err
	}

	res := &pb.CreateSnapshotResponse{
		Snapshotid: *result.SnapshotId,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	//First Creates the snapshot of the volume then deletes the volume

	logger := svc.logger

	volumeid := req.GetVolumeid()
	region := req.Region
	tags := addAWSTags(req.GetLabels())

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
	session, err := NewSession(svc.config, region, req.AccountName)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	ec2Client := session.getEC2Client()
	//calling aws sdk CreateSnapshot method
	resultSnapshot, err := ec2Client.CreateSnapshot(inputSnapshot)
	if err != nil {
		LogError("CreateSnapshot", logger, err)
		return &pb.CreateSnapshotAndDeleteResponse{}, err
	}

	err = ec2Client.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{resultSnapshot.SnapshotId},
	})
	if err != nil {
		LogError("WaitForSnapshotCompleted", logger, err)
		return &pb.CreateSnapshotAndDeleteResponse{}, err
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
		LogError("DeleteVolume", logger, err)

		return &pb.CreateSnapshotAndDeleteResponse{
			Snapshotid: *resultSnapshot.SnapshotId,
			Deleted:    false,
		}, err
	}

	//note: since now err is nil so assigning deleted = true
	res := &pb.CreateSnapshotAndDeleteResponse{
		Snapshotid: *resultSnapshot.SnapshotId,
		Deleted:    true,
	}

	return res, nil
}
