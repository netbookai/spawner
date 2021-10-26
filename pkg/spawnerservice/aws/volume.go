package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateAwsSession(provider string, region string, sessionName string, logger *zap.SugaredLogger) (awsSvc *ec2.EC2, err error) {
	//starts an AWS session
	accessKey, secretID, sessiontoken, stserr := GetCredsFromSTS(sessionName)
	if stserr != nil {
		logger.Errorw("Error getting Credentials")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretID, sessiontoken),
	})

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}
	awsSvc = ec2.New(sess)

	return awsSvc, err
}

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
	sessionName := "AWS create volume sesion, at " + time.Stamp
	logger := svc.logger

	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc, err := CreateAwsSession(provider, region, sessionName, logger)
	if err != nil {
		svc.logger.Errorw("error creating AWS session", "provider", provider, "region", region, "createvolrequest", req, "error", err)
		return &pb.CreateVolumeResponse{}, status.Errorf(codes.Internal, "error creating AWS session")
	}

	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availabilityZone),
		VolumeType:       aws.String(volumeType),
		Size:             aws.Int64(size),
		SnapshotId:       aws.String(snapshotId),
	}

	result, err := awsSvc.CreateVolume(input)
	LogError("CreateVolume", logger, err)

	res := &pb.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
	}
	return res, nil
}

func (svc AWSController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	//Deletes an EBS volume
	sessionName := "AWS delete volume session, at " + time.Stamp
	logger := svc.logger

	volumeid := req.GetVolumeid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc, err := CreateAwsSession(provider, region, sessionName, logger)
	if err != nil {
		svc.logger.Errorw("error creating AWS session", "provider", provider, "region", region, "createvolrequest", req, "error", err)
		return &pb.DeleteVolumeResponse{}, status.Errorf(codes.Internal, "error creating AWS session")
	}

	deleted, err := DeleteVolumeInternal(volumeid, awsSvc)

	LogError("DeleteVolume", logger, err)

	res := &pb.DeleteVolumeResponse{
		Deleted: deleted,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	//Creates a Snapshot of a volume
	sessionName := "AWS create snapshot session, at " + time.Stamp
	logger := svc.logger
	volumeid := req.GetVolumeid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc, err := CreateAwsSession(provider, region, sessionName, logger)
	if err != nil {
		svc.logger.Errorw("error creating AWS session", "provider", provider, "region", region, "createvolrequest", req, "error", err)
		return &pb.CreateSnapshotResponse{}, status.Errorf(codes.Internal, "error creating AWS session")
	}

	snapshotid, err := SnapshotInternal(volumeid, awsSvc)

	LogError("CreateSnapshot", logger, err)

	res := &pb.CreateSnapshotResponse{
		Snapshotid: snapshotid,
	}
	return res, nil
}

func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	//First Creates the snapshot of the volume then deletes the volume
	sessionName := "AWS create snapshot and delete session, at " + time.Stamp
	logger := svc.logger

	volumeid := req.GetVolumeid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc, err := CreateAwsSession(provider, region, sessionName, logger)
	if err != nil {
		svc.logger.Errorw("error creating AWS session", "provider", provider, "region", region, "createvolrequest", req, "error", err)
		return &pb.CreateSnapshotAndDeleteResponse{}, status.Errorf(codes.Internal, "error creating AWS session")
	}

	snapshotid, errS := SnapshotInternal(volumeid, awsSvc)

	deleted, errD := DeleteVolumeInternal(volumeid, awsSvc)

	if errS != nil || errD != nil {
		logger.Errorw("error in ", "method ", " CreateSnapshotAndDelete")

		if errS != nil {
			LogError("CreateSnapshot", logger, errS)
		}

		if errD != nil {
			LogError("DeleteVolume", logger, errD)
		}

	}

	res := &pb.CreateSnapshotAndDeleteResponse{
		Snapshotid: snapshotid,
		Deleted:    deleted,
	}
	return res, nil
}

func DeleteVolumeInternal(volumeid string, awsSvc *ec2.EC2) (deleted bool, err error) {
	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	_, err = awsSvc.DeleteVolume(input)
	deleted = true

	return deleted, err
}

func SnapshotInternal(volumeid string, awsSvc *ec2.EC2) (snapshotid string, err error) {
	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	result, err := awsSvc.CreateSnapshot(input)

	return *result.SnapshotId, err
}
