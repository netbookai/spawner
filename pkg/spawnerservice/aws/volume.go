package aws

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pb"
	grpc "google.golang.org/grpc"
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
	fmt.Println("Starting DeleteVol()")

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	awsSvc := ec2.New(sess)

	if err != nil {
		log.Fatalf("error starting aws session")
	}

	volumeid := req.GetVolumeid()
	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeid),
	}

	_, err = awsSvc.DeleteVolume(input)
	deleted := true
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
	res := &pb.DeleteVolRes{
		Deleted: deleted,
	}

	return res, nil
}

func main() {
	fmt.Println("hello world")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSpawnerServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
