package spwntransport

import (
	"context"
	"time"

	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	"google.golang.org/grpc"

	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

type grpcServer struct {
	createCluster grpctransport.Handler
	clusterStatus grpctransport.Handler
	addNode       grpctransport.Handler
	deleteCluster grpctransport.Handler
	deleteNode    grpctransport.Handler

	pb.UnimplementedSpawnerServiceServer
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints spwnendpoint.Set, logger log.Logger) pb.SpawnerServiceServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	return &grpcServer{
		createCluster: grpctransport.NewServer(
			endpoints.CreateClusterEndpoint,
			decodeGRPCClusterRequest,
			encodeGRPCClusterResponse,
			append(options)...,
		),
		clusterStatus: grpctransport.NewServer(
			endpoints.CusterStatusEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
		addNode: grpctransport.NewServer(
			endpoints.AddNodeEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
		deleteCluster: grpctransport.NewServer(
			endpoints.DeleteClusterEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
		deleteNode: grpctransport.NewServer(
			endpoints.DeleteNodeEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
	}
}

func (s *grpcServer) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	_, rep, err := s.createCluster.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ClusterResponse), nil
}

func (s *grpcServer) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	_, rep, err := s.clusterStatus.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ClusterStatusResponse), nil
}

func (s *grpcServer) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	_, rep, err := s.addNode.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.NodeSpawnResponse), nil
}

func (s *grpcServer) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	_, rep, err := s.deleteCluster.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ClusterDeleteResponse), nil
}

func (s *grpcServer) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	_, rep, err := s.deleteNode.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.NodeDeleteResponse), nil
}

// NewGRPCClient returns an AddService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) spawnerservice.ClusterController {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	// global client middlewares
	var options []grpctransport.ClientOption

	// Each individual endpoint is an grpc/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var createClusterEndpoint endpoint.Endpoint
	{
		createClusterEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"CreateCluster",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.ClusterResponse{},
			append(options)...,
		).Endpoint()
		createClusterEndpoint = limiter(createClusterEndpoint)
		createClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateCluster",
			Timeout: 30 * time.Second,
		}))(createClusterEndpoint)
	}

	var clusterStatusEndpoint endpoint.Endpoint
	{
		clusterStatusEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"ClusterStatus",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.ClusterStatusResponse{},
			append(options)...,
		).Endpoint()
		clusterStatusEndpoint = limiter(clusterStatusEndpoint)
		clusterStatusEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "ClusterStatus",
			Timeout: 30 * time.Second,
		}))(clusterStatusEndpoint)
	}

	var addNodeEndpoint endpoint.Endpoint
	{
		addNodeEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"AddNode",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.NodeSpawnResponse{},
			append(options)...,
		).Endpoint()
		addNodeEndpoint = limiter(addNodeEndpoint)
		addNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "AddNode",
			Timeout: 30 * time.Second,
		}))(addNodeEndpoint)
	}

	var deleteClusterEndpoint endpoint.Endpoint
	{
		deleteClusterEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"DeleteCluster",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.ClusterDeleteResponse{},
			append(options)...,
		).Endpoint()
		deleteClusterEndpoint = limiter(deleteClusterEndpoint)
		deleteClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "DeleteCluster",
			Timeout: 30 * time.Second,
		}))(deleteClusterEndpoint)
	}

	var deleteNodeEndpoint endpoint.Endpoint
	{
		deleteNodeEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"DeleteNode",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.NodeDeleteResponse{},
			append(options)...,
		).Endpoint()
		deleteNodeEndpoint = limiter(deleteNodeEndpoint)
		deleteNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "DeleteNode",
			Timeout: 30 * time.Second,
		}))(deleteNodeEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return spwnendpoint.Set{
		CreateClusterEndpoint: createClusterEndpoint,
		CusterStatusEndpoint:  clusterStatusEndpoint,
		AddNodeEndpoint:       addNodeEndpoint,
		DeleteClusterEndpoint: deleteClusterEndpoint,
		DeleteNodeEndpoint:    deleteNodeEndpoint,
	}
}

// decodeGRPCSumRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func decodeGRPCClusterRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	// req := grpcReq.(*pb.ClusterRequest)
	// return req, nil
	return grpcReq, nil
}

// encodeGRPCSumResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain sum response to a gRPC sum reply. Primarily useful in a server.
func encodeGRPCClusterResponse(_ context.Context, response interface{}) (interface{}, error) {
	// resp := response.(*pb.ClusterResponse)
	// return &resp, nil
	return response, nil
}
