package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pb"
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

func (svc AWSController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	//Creates an EBS volume

	logger := svc.logger

	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()

	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availabilityZone),
		VolumeType:       aws.String(volumeType),
		Size:             aws.Int64(size),
		SnapshotId:       aws.String(snapshotId),
	}

	//calling aws sdk method to create volume
	result, err := svc.client.CreateVolume(input)

	LogError("CreateVolume", logger, err)

	if err != nil {
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

	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to delete volume
	//ec2.DeleteVolumeOutput doesn't contain anything
	//hence not taking response
	_, err := svc.client.DeleteVolume(input)

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

	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to snapshot volume
	result, err := svc.client.CreateSnapshot(input)

	if err != nil {
		LogError("CreateSnapshot", logger, err)
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

	inputSnapshot := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to snapshot volume
	resultSnapshot, err := svc.client.CreateSnapshot(inputSnapshot)

	if err != nil {
		LogError("CreateSnapshot", logger, err)
		return &pb.CreateSnapshotAndDeleteResponse{}, err
	}

	inputDelete := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	//calling aws sdk method to delete volume
	_, err = svc.client.DeleteVolume(inputDelete)

	if err != nil {
		LogError("DeleteVolume", logger, err)

		return &pb.CreateSnapshotAndDeleteResponse{
			Snapshotid: *resultSnapshot.SnapshotId,
			Deleted:    false,
		}, err
	}

	res := &pb.CreateSnapshotAndDeleteResponse{
		Snapshotid: *resultSnapshot.SnapshotId,
		Deleted:    true,
	}

	return res, nil
}
