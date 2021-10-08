package spwnendpoint

import (
	"context"
	"fmt"
	"time"

	// stdopentracing "github.com/opentracing/opentracing-go"
	// stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"

	// "github.com/go-kit/kit/log"
	// "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	// "github.com/go-kit/kit/tracing/opentracing"
	// "github.com/go-kit/kit/tracing/zipkin"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
)

// Set collects all of the endpoints that compose an add service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type Set struct {
	CreateClusterEndpoint           endpoint.Endpoint
	CusterStatusEndpoint            endpoint.Endpoint
	AddNodeEndpoint                 endpoint.Endpoint
	DeleteClusterEndpoint           endpoint.Endpoint
	DeleteNodeEndpoint              endpoint.Endpoint
	CreateVolEndpoint               endpoint.Endpoint
	DeleteVolEndpoint               endpoint.Endpoint
	CreateSnapshotEndpoint          endpoint.Endpoint
	CreateSnapshotAndDeleteEndpoint endpoint.Endpoint
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc spawnerservice.ClusterController) Set {
	var createClusterEndpoint endpoint.Endpoint
	{
		createClusterEndpoint = MakeCreateClusterEndpoint(svc)
		// Sum is limited to 1 request per second with burst of 1 request.
		// Note, rate is defined as a time interval between requests.
		createClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(createClusterEndpoint)
		createClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createClusterEndpoint)
	}

	var clusterStatusEndpoint endpoint.Endpoint
	{
		clusterStatusEndpoint = MakeCusterStatusEndpoint(svc)
		clusterStatusEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(clusterStatusEndpoint)
		clusterStatusEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(clusterStatusEndpoint)
	}

	var addNodeEndpoint endpoint.Endpoint
	{
		addNodeEndpoint = MakeAddNodeEndpoint(svc)
		addNodeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(addNodeEndpoint)
		addNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addNodeEndpoint)
	}

	var deleteClusterEndpoint endpoint.Endpoint
	{
		deleteClusterEndpoint = MakeClusterDeleteEndpoint(svc)
		deleteClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(deleteClusterEndpoint)
		deleteClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteClusterEndpoint)
	}

	var deleteNodeEndpoint endpoint.Endpoint
	{
		deleteNodeEndpoint = MakeNodeDeleteEndpoint(svc)
		deleteNodeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(deleteNodeEndpoint)
		deleteNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteNodeEndpoint)
	}

	var createVolEndpoint endpoint.Endpoint
	{
		createVolEndpoint = MakeCreateVolEndpoint(svc)
		createVolEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(createVolEndpoint)
		createVolEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createVolEndpoint)
	}

	var deleteVolEndpoint endpoint.Endpoint
	{
		deleteVolEndpoint = MakeDeleteVolEndpoint(svc)
		deleteVolEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(deleteVolEndpoint)
		deleteVolEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteVolEndpoint)
	}

	var createSnapshotEndpoint endpoint.Endpoint
	{
		createSnapshotEndpoint = MakeCreateSnapshotEndpoint(svc)
		createSnapshotEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(createSnapshotEndpoint)
		createSnapshotEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createSnapshotEndpoint)
	}

	var createSnapshotAndDeleteEndpoint endpoint.Endpoint
	{
		createSnapshotAndDeleteEndpoint = MakeCreateSnapshotAndDeleteEndpoint(svc)
		createSnapshotAndDeleteEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createSnapshotAndDeleteEndpoint)
	}

	return Set{
		CreateClusterEndpoint:           createClusterEndpoint,
		CusterStatusEndpoint:            clusterStatusEndpoint,
		AddNodeEndpoint:                 addNodeEndpoint,
		DeleteClusterEndpoint:           deleteClusterEndpoint,
		DeleteNodeEndpoint:              deleteNodeEndpoint,
		CreateVolEndpoint:               createVolEndpoint,
		DeleteVolEndpoint:               deleteVolEndpoint,
		CreateSnapshotEndpoint:          createSnapshotEndpoint,
		CreateSnapshotAndDeleteEndpoint: createSnapshotAndDeleteEndpoint,
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	resp, err := s.CreateClusterEndpoint(ctx, req)
	if err != nil {
		return &pb.ClusterResponse{}, err
	}
	response := resp.(*pb.ClusterResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeCreateClusterEndpoint constructs a CreateCluster endpoint wrapping the service.
func MakeCreateClusterEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.ClusterRequest)
		resp, err := s.CreateCluster(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	resp, err := s.CusterStatusEndpoint(ctx, req)
	if err != nil {
		return &pb.ClusterStatusResponse{}, err
	}
	response := resp.(*pb.ClusterStatusResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeCusterStatusEndpoint constructs a ClusterStatus endpoint wrapping the service.
func MakeCusterStatusEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.ClusterStatusRequest)
		resp, err := s.ClusterStatus(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	resp, err := s.AddNodeEndpoint(ctx, req)
	if err != nil {
		return &pb.NodeSpawnResponse{}, err
	}
	response := resp.(*pb.NodeSpawnResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeAddNodeEndpoint constructs a AddNode endpoint wrapping the service.
func MakeAddNodeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.NodeSpawnRequest)
		resp, err := s.AddNode(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	resp, err := s.DeleteClusterEndpoint(ctx, req)
	if err != nil {
		return &pb.ClusterDeleteResponse{}, err
	}
	response := resp.(*pb.ClusterDeleteResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeClusterDeleteEndpointt constructs a ClusterStatus endpoint wrapping the service.
func MakeClusterDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.ClusterDeleteRequest)
		resp, err := s.DeleteCluster(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	resp, err := s.DeleteNodeEndpoint(ctx, req)
	if err != nil {
		return &pb.NodeDeleteResponse{}, err
	}
	response := resp.(*pb.NodeDeleteResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeClusterDeleteEndpointt constructs a ClusterStatus endpoint wrapping the service.
func MakeNodeDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.NodeDeleteRequest)
		resp, err := s.DeleteNode(ctx, req)
		return resp, err
	}
}

func (s Set) CreateVol(ctx context.Context, req *pb.CreateVolReq) (*pb.CreateVolRes, error) {
	resp, err := s.CreateVolEndpoint(ctx, req)
	if err != nil {
		return &pb.CreateVolRes{}, err
	}
	response := resp.(*pb.CreateVolRes)
	// TODO: Shivani add error to CreateVolRes and use it here
	return response, fmt.Errorf("")
}

func MakeCreateVolEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.CreateVolReq)
		resp, err := s.CreateVol(ctx, req)
		return resp, err
	}
}

func (s Set) DeleteVol(ctx context.Context, req *pb.DeleteVolReq) (*pb.DeleteVolRes, error) {
	resp, err := s.DeleteVolEndpoint(ctx, req)
	if err != nil {
		return &pb.DeleteVolRes{}, err
	}
	response := resp.(*pb.DeleteVolRes)
	// TODO: Shivani add error to CreateVolRes and use it here
	return response, fmt.Errorf("")
}

func MakeDeleteVolEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.DeleteVolReq)
		resp, err := s.DeleteVol(ctx, req)
		return resp, err
	}
}

func (s Set) CreateSnapshot(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {
	resp, err := s.CreateSnapshotEndpoint(ctx, req)
	if err != nil {
		return &pb.SnapshotResponse{}, err
	}
	response := resp.(*pb.SnapshotResponse)
	// TODO: Shivani add error to CreateVolRes and use it here
	return response, fmt.Errorf("")
}

func MakeCreateSnapshotEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SnapshotRequest)
		resp, err := s.CreateSnapshot(ctx, req)
		return resp, err
	}
}

func (s Set) CreateSnapshotAndDelete(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {
	resp, err := s.CreateSnapshotAndDeleteEndpoint(ctx, req)
	if err != nil {
		return &pb.SnapshotResponse{}, err
	}
	response := resp.(*pb.SnapshotResponse)
	// TODO: Shivani add error to CreateVolRes and use it here
	return response, fmt.Errorf("")
}

func MakeCreateSnapshotAndDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SnapshotRequest)
		resp, err := s.CreateSnapshotAndDelete(ctx, req)
		return resp, err
	}
}
