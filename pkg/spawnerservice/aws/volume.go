package aws

import (
	"context"
	"fmt"
	"log"

	//"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pb"
	//grpc "google.golang.org/grpc"
)

func CreateAwsSession(provider string, region string) (awsSvc *ec2.EC2) {
	//starts an AWS session

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	awsSvc = ec2.New(sess)
	if err != nil {
		log.Fatalf("error starting aws session")
	}
	return awsSvc
}

func (svc AWSController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	//Creates an EBS volume
	fmt.Println("CreateVol() invoked with ", req)

	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc := CreateAwsSession(provider, region)

	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availabilityZone),
		VolumeType:       aws.String(volumeType),
		Size:             aws.Int64(size),
		SnapshotId:       aws.String(snapshotId),
	}

	result, err := awsSvc.CreateVolume(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		log.Fatalf("failed to create vol: %v", err)
	}

	res := &pb.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
	}
	return res, nil
}

func (svc AWSController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	//Deletes an EBS volume
	fmt.Println("DeleteVol() invoked with ", req)

	volumeid := req.GetVolumeid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc := CreateAwsSession(provider, region)

	deleted := DeleteVolumeInternal(volumeid, awsSvc)

	res := &pb.DeleteVolumeResponse{
		Deleted: deleted,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	//Creates a Snapshot of a volume
	fmt.Println("CreateSnapshot() invoked with ", req)

	volumeid := req.GetVolumeid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc := CreateAwsSession(provider, region)

	snapshotid := SnapshotInternal(volumeid, awsSvc)

	res := &pb.CreateSnapshotResponse{
		Snapshotid: snapshotid,
	}
	return res, nil
}

func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	//First Creates the snapshot of the volume then deletes the volume
	fmt.Println("CreateSnapshotAndDelete() invoked with ", req)

	volumeid := req.GetVolumeid()
	provider := req.GetProvider()
	region := req.GetRegion()

	awsSvc := CreateAwsSession(provider, region)

	snapshotid := SnapshotInternal(volumeid, awsSvc)
	deleted := DeleteVolumeInternal(volumeid, awsSvc)

	res := &pb.CreateSnapshotAndDeleteResponse{
		Snapshotid: snapshotid,
		Deleted:    deleted,
	}
	return res, nil
}

func DeleteVolumeInternal(volumeid string, awsSvc *ec2.EC2) (deleted bool) {
	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	_, err := awsSvc.DeleteVolume(input)
	deleted = true
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		deleted = false
		log.Fatalf("error in deleting volume")
	}
	return deleted
}

func SnapshotInternal(volumeid string, awsSvc *ec2.EC2) (snapshotid string) {
	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	result, err := awsSvc.CreateSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		log.Fatalf("couldn't create snapshot ", err)
	}
	return *result.SnapshotId
}
