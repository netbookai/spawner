package aws

//import (
//	"context"
//	"testing"
//
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/service/ec2"
//	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
//	"gitlab.com/netbook-devs/spawner-service/proto"
//)
//
//type mockedCreateVolume struct {
//	ec2iface.EC2API
//	Resp             ec2.Volume
//	CreateVolumeMock func(in *ec2.CreateVolumeInput) (*ec2.Volume, error)
//}
//
//func (c mockedCreateVolume) CreateVolume(in *ec2.CreateVolumeInput) (*ec2.Volume, error) {
//	return c.CreateVolumeMock(in)
//}
//
//type mockedDeleteVolume struct {
//	ec2iface.EC2API
//	Resp             ec2.DeleteVolumeOutput
//	DeleteVolumeMock func(in *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
//}
//
//func (c mockedDeleteVolume) DeleteVolume(in *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
//	return c.DeleteVolumeMock(in)
//}
//
//type mockedCreateSnapshot struct {
//	ec2iface.EC2API
//	Resp               ec2.Snapshot
//	CreateSnapshotMock func(in *ec2.CreateSnapshotInput) (*ec2.Snapshot, error)
//}
//
//func (c mockedCreateSnapshot) CreateSnapshot(in *ec2.CreateSnapshotInput) (*ec2.Snapshot, error) {
//	return c.CreateSnapshotMock(in)
//}
//
//type mockedCreateSnapshotAndDelete struct {
//	ec2iface.EC2API
//	ResponseSnapshot   ec2.Snapshot
//	ResponseDelete     ec2.DeleteVolumeOutput
//	CreateSnapshotMock func(in *ec2.CreateSnapshotInput) (*ec2.Snapshot, error)
//	DeleteVolumeMock   func(in *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error)
//}
//
//func (c mockedCreateSnapshotAndDelete) CreateSnapshot(in *ec2.CreateSnapshotInput) (*ec2.Snapshot, error) {
//	return c.CreateSnapshotMock(in)
//}
//
//func (c mockedCreateSnapshotAndDelete) DeleteVolume(in *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
//	return c.DeleteVolumeMock(in)
//}
//
//func TestCreateVolume(t *testing.T) {
//
//	type testCase struct {
//		name           string                   // test case name
//		input          *proto.CreateVolumeRequest  // function input
//		expectedOutput *proto.CreateVolumeResponse // expected outcome
//		err            error                    //error
//	}
//
//	//CreateVolume() test input and expected outputs
//	var createVolumeRequest1 = &proto.CreateVolumeRequest{
//		Availabilityzone: "us-west-2a",
//		Volumetype:       "gp2",
//		Size:             1,
//		Snapshotid:       "",
//		Provider:         "aws",
//		Region:           "us-west-2",
//	}
//
//	var createVolumeResponse1 = &proto.CreateVolumeResponse{
//		Volumeid: *aws.String("test-vol-id"),
//		Error:    "",
//	}
//
//	testTable := []testCase{
//		{
//			name:           "all correct input",
//			input:          createVolumeRequest1,
//			expectedOutput: createVolumeResponse1,
//			err:            nil,
//		},
//	}
//
//	// Begin test
//	for _, test := range testTable {
//
//		v := AWSController{
//			ec2SessFactory: func(region string) (ec2iface.EC2API, error) {
//				m := mockedCreateVolume{
//
//					Resp: ec2.Volume{
//						VolumeId: &test.expectedOutput.Volumeid,
//					},
//
//					CreateVolumeMock: func(in *ec2.CreateVolumeInput) (*ec2.Volume, error) {
//						return &ec2.Volume{
//							VolumeId: aws.String(createVolumeResponse1.Volumeid),
//						}, nil
//					},
//				}
//				return m, nil
//			},
//		}
//
//		//calling the function to be tested
//		actual, actualerr := v.CreateVolume(context.Background(), test.input)
//
//		//asserting
//		if actual.Volumeid != test.expectedOutput.Volumeid {
//			t.Errorf("expected volume id is %v, found %v", test.expectedOutput.Volumeid, actual.Volumeid)
//		}
//
//		if actualerr != nil && actual.Volumeid != "" {
//			t.Errorf("volume id should be empty, found %s", actual.Volumeid)
//		}
//
//		if actualerr == nil && actual.Volumeid == "" {
//			t.Errorf("volume id should be non empty, found volume id %s", actual.Volumeid)
//		}
//
//		if actual.Error != test.expectedOutput.Error {
//			t.Errorf("expected error field %s, got error field %s", test.expectedOutput.Error, actual.Error)
//		}
//
//		if actualerr != test.err {
//			t.Errorf(" expected error %v, got error %v", test.err, actualerr)
//		}
//
//	}
//}
//
//func TestDeleteVolume(t *testing.T) {
//
//	type testCase struct {
//		name           string                   // test case name
//		input          *proto.DeleteVolumeRequest  // function input
//		expectedOutput *proto.DeleteVolumeResponse // expected outcome
//		err            error                    //error
//	}
//
//	//DeleteVolume() test inputs and expected outputs
//	var deleteVolumeRequest1 = &proto.DeleteVolumeRequest{
//		Volumeid: *aws.String("test-vol-id"),
//		Provider: "aws",
//		Region:   "us-west-2",
//	}
//
//	var deleteVolumeResponse1 = &proto.DeleteVolumeResponse{
//		Deleted: true,
//		Error:   "",
//	}
//
//	testTable := []testCase{
//		{
//			name:           "all correct input",
//			input:          deleteVolumeRequest1,
//			expectedOutput: deleteVolumeResponse1,
//		},
//	}
//
//	// Begin test
//	for _, test := range testTable {
//
//		v := AWSController{
//			ec2SessFactory: func(region string) (ec2iface.EC2API, error) {
//				m := mockedDeleteVolume{
//
//					Resp: ec2.DeleteVolumeOutput{},
//
//					DeleteVolumeMock: func(in *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
//						return &ec2.DeleteVolumeOutput{}, nil
//					},
//				}
//				return m, nil
//			},
//		}
//
//		//calling the function to be tested
//		actual, actualerr := v.DeleteVolume(context.Background(), test.input)
//
//		//asserting
//		if actualerr != nil && actual.Deleted == true {
//			t.Errorf("Volume shouldn't be deleted, Value of 'deleted' field is, %t", actual.Deleted)
//		}
//
//		if actualerr == nil && actual.Deleted == false {
//			t.Errorf("Volume should be deleted, Value of 'deleted' field is, %t", actual.Deleted)
//		}
//
//		if actual.Error != test.expectedOutput.Error {
//			t.Errorf("expected error field %s, got error field %s", test.expectedOutput.Error, actual.Error)
//		}
//
//		if actualerr != test.err {
//			t.Errorf(" expected error %v, got error %v", test.err, actualerr)
//		}
//
//	}
//}
//func TestCreateSnapshot(t *testing.T) {
//
//	type testCase struct {
//		name           string                     // test case name
//		input          *proto.CreateSnapshotRequest  // function input
//		expectedOutput *proto.CreateSnapshotResponse // expected outcome
//		err            error                      //error
//	}
//
//	//CreateSnapshot() test inputs and expected outputs
//	var createSnapshotRequest1 = &proto.CreateSnapshotRequest{
//		Volumeid: "test-vol-id",
//		Provider: "aws",
//		Region:   "us-west-2",
//	}
//
//	var createSnapshotResponse1 = &proto.CreateSnapshotResponse{
//		Snapshotid: "test-snapshot-id",
//		Error:      "",
//	}
//	testTable := []testCase{
//		{
//			name:           "all correct input",
//			input:          createSnapshotRequest1,
//			expectedOutput: createSnapshotResponse1,
//		},
//	}
//
//	// Begin test
//
//	for _, test := range testTable {
//
//		v := AWSController{
//			ec2SessFactory: func(region string) (ec2iface.EC2API, error) {
//				m := mockedCreateSnapshot{
//
//					Resp: ec2.Snapshot{
//						SnapshotId: &test.expectedOutput.Snapshotid,
//					},
//
//					CreateSnapshotMock: func(in *ec2.CreateSnapshotInput) (*ec2.Snapshot, error) {
//						return &ec2.Snapshot{
//							SnapshotId: aws.String(createSnapshotResponse1.Snapshotid),
//						}, nil
//					},
//				}
//				return m, nil
//			},
//		}
//
//		//calling the function to be tested
//		actual, actualerr := v.CreateSnapshot(context.Background(), test.input)
//
//		//asserting
//		if actualerr != nil && actual.Snapshotid != "" {
//			t.Errorf("snapshot id should be empty, found %s", actual.Snapshotid)
//		}
//
//		if actualerr == nil && actual.Snapshotid == "" {
//			t.Errorf("snapshot id should be non empty, found volume id %s", actual.Snapshotid)
//		}
//
//		if actual.Error != test.expectedOutput.Error {
//			t.Errorf("expected error field %s, got error field %s", test.expectedOutput.Error, actual.Error)
//		}
//
//		if actualerr != test.err {
//			t.Errorf(" expected error %v, got error %v", test.err, actualerr)
//		}
//
//	}
//}
//
//func TestCreateSnapshotAndDelete(t *testing.T) {
//
//	type testCase struct {
//		name           string                              // test case name
//		input          *proto.CreateSnapshotAndDeleteRequest  // function input
//		expectedOutput *proto.CreateSnapshotAndDeleteResponse // expected outcome
//		err            error                               //error
//	}
//
//	// //CreateSnapshotAndDelete() test inputs and expected outputs
//	var createSnapshotAndDeleteRequest1 = &proto.CreateSnapshotAndDeleteRequest{
//		Volumeid: *aws.String("test-vol-id"),
//		Provider: "aws",
//		Region:   "us-west-2",
//	}
//
//	var createSnapshotAndDeleteResponse1 = &proto.CreateSnapshotAndDeleteResponse{
//		Snapshotid: "test-snapshot-id",
//		Deleted:    true,
//		Error:      "",
//	}
//
//	testTable := []testCase{
//		{
//			name:           "all correct input",
//			input:          createSnapshotAndDeleteRequest1,
//			expectedOutput: createSnapshotAndDeleteResponse1,
//		},
//	}
//
//	// Begin test
//	for _, test := range testTable {
//
//		v := AWSController{
//			ec2SessFactory: func(region string) (ec2iface.EC2API, error) {
//				m := mockedCreateSnapshotAndDelete{
//
//					ResponseSnapshot: ec2.Snapshot{
//						SnapshotId: &test.expectedOutput.Snapshotid,
//					},
//
//					ResponseDelete: ec2.DeleteVolumeOutput{},
//
//					CreateSnapshotMock: func(in *ec2.CreateSnapshotInput) (*ec2.Snapshot, error) {
//						return &ec2.Snapshot{
//							SnapshotId: aws.String(createSnapshotAndDeleteResponse1.Snapshotid),
//						}, nil
//					},
//
//					DeleteVolumeMock: func(in *ec2.DeleteVolumeInput) (*ec2.DeleteVolumeOutput, error) {
//						return &ec2.DeleteVolumeOutput{}, nil
//					},
//				}
//				return m, nil
//			},
//		}
//
//		//calling the function to be tested
//		createSnapshotAndDeleteInput := proto.CreateSnapshotAndDeleteRequest{
//			Volumeid: test.input.Volumeid,
//		}
//		actual, actualerr := v.CreateSnapshotAndDelete(context.Background(), &createSnapshotAndDeleteInput)
//
//		//asserting
//		if actualerr != nil && actual.Snapshotid != "" {
//			t.Errorf("snapshot id should be empty, found %s", actual.Snapshotid)
//		}
//
//		if actualerr == nil && actual.Snapshotid == "" {
//			t.Errorf("snapshot id should be non empty, found volume id %s", actual.Snapshotid)
//		}
//
//		if actual.Deleted == false && actual.Snapshotid != "" {
//			t.Errorf("if volume is not deleted, snapshot id should be empty. SNapshot id is %s", actual.Snapshotid)
//		}
//
//		if actual.Deleted == true && actual.Snapshotid == "" {
//			t.Errorf("if volume is deleted, snapshot id should be non empty. Snapshot id is %s", actual.Snapshotid)
//		}
//
//		if actual.Error != test.expectedOutput.Error {
//			t.Errorf("expected error field %s, got error field %s", test.expectedOutput.Error, actual.Error)
//		}
//
//		if actualerr != test.err {
//			t.Errorf(" expected error %v, got error %v", test.err, actualerr)
//		}
//
//	}
//}
