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
	createCluster           grpctransport.Handler
	addToken                grpctransport.Handler
	getToken                grpctransport.Handler
	clusterStatus           grpctransport.Handler
	addNode                 grpctransport.Handler
	deleteCluster           grpctransport.Handler
	deleteNode              grpctransport.Handler
	createVolume            grpctransport.Handler
	deleteVolume            grpctransport.Handler
	createSnapshot          grpctransport.Handler
	createSnapshotAndDelete grpctransport.Handler

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
		addToken: grpctransport.NewServer(
			endpoints.AddTokenEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
		getToken: grpctransport.NewServer(
			endpoints.GetTokenEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
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
		createVolume: grpctransport.NewServer(
			endpoints.CreateVolumeEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),

		deleteVolume: grpctransport.NewServer(
			endpoints.DeleteVolumeEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),

		createSnapshot: grpctransport.NewServer(
			endpoints.CreateSnapshotEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),

		createSnapshotAndDelete: grpctransport.NewServer(
			endpoints.CreateSnapshotAndDeleteEndpoint,
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

func (s *grpcServer) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	_, rep, err := s.addToken.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.AddTokenResponse), nil
}

func (s *grpcServer) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	_, rep, err := s.getToken.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.GetTokenResponse), nil
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

func (s *grpcServer) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	_, rep, err := s.createVolume.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateVolumeResponse), nil
}

func (s *grpcServer) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	_, rep, err := s.deleteVolume.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.DeleteVolumeResponse), nil
}

func (s *grpcServer) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	_, rep, err := s.createSnapshot.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateSnapshotResponse), nil
}

func (s *grpcServer) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	_, rep, err := s.createSnapshotAndDelete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateSnapshotAndDeleteResponse), nil
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

	var addTokenEndpoint endpoint.Endpoint
	{
		addTokenEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"AddToken",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.AddTokenResponse{},
			append(options)...,
		).Endpoint()
		addTokenEndpoint = limiter(addTokenEndpoint)
		addTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "AddToken",
			Timeout: 30 * time.Second,
		}))(addTokenEndpoint)
	}

	var getTokenEndpoint endpoint.Endpoint
	{
		getTokenEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"GetToken",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.GetTokenResponse{},
			append(options)...,
		).Endpoint()
		getTokenEndpoint = limiter(getTokenEndpoint)
		getTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetToken",
			Timeout: 30 * time.Second,
		}))(getTokenEndpoint)
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

	var createVolumeEndpoint endpoint.Endpoint
	{
		createVolumeEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"CreateVolume",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.CreateVolumeResponse{},
			append(options)...,
		).Endpoint()
		createVolumeEndpoint = limiter(createVolumeEndpoint)
		createVolumeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateVol",
			Timeout: 30 * time.Second,
		}))(createVolumeEndpoint)
	}

	var deleteVolumeEndpoint endpoint.Endpoint
	{
		deleteVolumeEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"DeleteVolume",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.DeleteVolumeResponse{},
			append(options)...,
		).Endpoint()
		deleteVolumeEndpoint = limiter(deleteVolumeEndpoint)
		deleteVolumeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "DeleteVol",
			Timeout: 30 * time.Second,
		}))(deleteVolumeEndpoint)
	}

	var createSnapshotEndpoint endpoint.Endpoint
	{
		createSnapshotEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"CreateSnapshot",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.CreateSnapshotResponse{},
			append(options)...,
		).Endpoint()
		createSnapshotEndpoint = limiter(createSnapshotEndpoint)
		createSnapshotEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateSnapshot",
			Timeout: 30 * time.Second,
		}))(createSnapshotEndpoint)
	}

	var createSnapshotAndDeleteEndpoint endpoint.Endpoint
	{
		createSnapshotAndDeleteEndpoint = grpctransport.NewClient(
			conn,
			"pb.SpawnerService",
			"CreateSnapshotAndDelete",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			pb.CreateSnapshotAndDeleteResponse{},
			append(options)...,
		).Endpoint()
		createSnapshotAndDeleteEndpoint = limiter(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateSnapshotAndDelete",
			Timeout: 30 * time.Second,
		}))(createSnapshotAndDeleteEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return spwnendpoint.Set{
		CreateClusterEndpoint:           createClusterEndpoint,
		AddTokenEndpoint:                addTokenEndpoint,
		GetTokenEndpoint:                getTokenEndpoint,
		CusterStatusEndpoint:            clusterStatusEndpoint,
		AddNodeEndpoint:                 addNodeEndpoint,
		DeleteClusterEndpoint:           deleteClusterEndpoint,
		DeleteNodeEndpoint:              deleteNodeEndpoint,
		CreateVolumeEndpoint:            createVolumeEndpoint,
		DeleteVolumeEndpoint:            deleteVolumeEndpoint,
		CreateSnapshotEndpoint:          createSnapshotEndpoint,
		CreateSnapshotAndDeleteEndpoint: createSnapshotAndDeleteEndpoint,
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
