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

type server struct {
	pb.UnimplementedSpawnerServiceServer
}

func (svc AWSController) CreateVol(ctx context.Context, req *pb.CreateVolReq) (*pb.CreateVolRes, error) {
	fmt.Println("CreateVol() invoked with ", req)
	//getting input values
	availabilityZone := req.GetAvailabilityzone()
	volumeType := req.GetVolumetype()
	size := req.GetSize()
	snapshotId := req.GetSnapshotid()

	//staring aws session
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	awsSvc := ec2.New(sess)

	if err != nil {
		log.Fatalf("error starting aws session")
	}

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

	res := &pb.CreateVolRes{
		Volumeid: *result.VolumeId,
	}
	return res, nil
}

func (svc AWSController) DeleteVol(ctx context.Context, req *pb.DeleteVolReq) (*pb.DeleteVolRes, error) {
	fmt.Println("DeleteVol() invoked with ", req)

	volumeid := req.GetVolumeid()
	deleted := DeleteVolInternal(volumeid)

	res := &pb.DeleteVolRes{
		Deleted: deleted,
	}

	return res, nil
}

func (svc AWSController) CreateSnapshot(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {

	fmt.Println("CreateSnapshot() invoked with ", req)
	volumeid := req.GetVolumeid()
	snapshotid, _ := SnapshotInternal(volumeid, false)

	res := &pb.SnapshotResponse{
		Snapshotid: snapshotid,
	}
	return res, nil
}

func (svc AWSController) CreateSnapshotAndDelete(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {

	fmt.Println("CreateSnapshotAndDelete() invoked with ", req)
	volumeid := req.GetVolumeid()
	snapshotid, _ := SnapshotInternal(volumeid, true)

	res := &pb.SnapshotResponse{
		Snapshotid: snapshotid,
	}
	return res, nil
}

func DeleteVolInternal(volumeid string) (deleted bool) {
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
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	svc := ec2.New(sess)

	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeid),
	}

	result, err := svc.CreateSnapshot(input)
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
			deleted = DeleteVolInternal(volumeid)
		}
	}
	return *result.SnapshotId, deleted
}

// func main() {
// 	fmt.Println("hello world")

// 	lis, err := net.Listen("tcp", "0.0.0.0:50051")

// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}

// 	s := grpc.NewServer()
// 	pb.RegisterSpawnerServiceServer(s, &server{})

// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}

// }
