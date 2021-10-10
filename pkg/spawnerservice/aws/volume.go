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

func CreateAwsSession(region string) (awsSvc *ec2.EC2) {

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	awsSvc = ec2.New(sess)
	if err != nil {
		log.Fatalf("error starting aws session")
	}
	return awsSvc
}

func (svc AWSController) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	fmt.Println("CreateVol() invoked with ", req)
	//getting input values
	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()

	//staring aws session
	awsSvc := CreateAwsSession("us-west-2")

	//assigning input values for CreateVolume()
	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availabilityZone),
		VolumeType:       aws.String(volumeType),
		Size:             aws.Int64(size),
		SnapshotId:       aws.String(snapshotId),
	}

	//creating volume
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
	//fmt.Println(result)

	res := &pb.CreateVolumeResponse{
		Volumeid: *result.VolumeId,
	}
	return res, nil
}

func (svc AWSController) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	fmt.Println("DeleteVol() invoked with ", req)

	volumeid := req.GetVolumeid()
	deleted := DeleteVolumeInternal(volumeid)

	res := &pb.DeleteVolumeResponse{
		Deleted: deleted,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshot(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {

	fmt.Println("CreateSnapshot() invoked with ", req)
	volumeid := req.GetVolumeid()
	deletevol := false
	snapshotid, _ := SnapshotInternal(volumeid, deletevol)

	res := &pb.SnapshotResponse{
		Snapshotid: snapshotid,
	}
	return res, nil
}

func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {

	fmt.Println("CreateSnapshotAndDelete() invoked with ", req)
	volumeid := req.GetVolumeid()
	deletevol := true
	snapshotid, _ := SnapshotInternal(volumeid, deletevol)

	res := &pb.SnapshotResponse{
		Snapshotid: snapshotid,
	}
	return res, nil
}

func DeleteVolumeInternal(volumeid string) (deleted bool) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	awsSvc := ec2.New(sess)

	if err != nil {
		log.Fatalf("error starting aws session")
	}

	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	_, err = awsSvc.DeleteVolume(input)
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
		log.Fatalf("error delete volume")
	}
	return deleted
}

func SnapshotInternal(volumeid string, deletevol bool) (snapshotid string, deleted bool) {
	//staring aws session
	awsSvc := CreateAwsSession("us-west-2")

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
	} else {
		if deletevol == true {
			deleted = DeleteVolumeInternal(volumeid)
		}
	}
	return *result.SnapshotId, deleted
}
