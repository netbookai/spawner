package spwntransport

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/transport"
	"github.com/sony/gobreaker"

	kitzap "github.com/go-kit/kit/log/zap"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type grpcServer struct {
	createCluster           grpctransport.Handler
	addToken                grpctransport.Handler
	getToken                grpctransport.Handler
	addRoute53Record        grpctransport.Handler
	getClusters             grpctransport.Server
	getCluster              grpctransport.Server
	clusterStatus           grpctransport.Handler
	addNode                 grpctransport.Handler
	deleteCluster           grpctransport.Handler
	deleteNode              grpctransport.Handler
	createVolume            grpctransport.Handler
	deleteVolume            grpctransport.Handler
	createSnapshot          grpctransport.Handler
	createSnapshotAndDelete grpctransport.Handler
	registerWithRancher     grpctransport.Handler

	proto.UnimplementedSpawnerServiceServer
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints spwnendpoint.Set, logger *zap.SugaredLogger) proto.SpawnerServiceServer {
	kitZapLogger := kitzap.NewZapSugarLogger(logger.Desugar(), zap.InfoLevel)
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(kitZapLogger)),
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
		addRoute53Record: grpctransport.NewServer(
			endpoints.AddRoute53RecordEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
		getClusters: *grpctransport.NewServer(
			endpoints.GetClustersEndpoint,
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, response interface{}) (interface{}, error) {
				return response, nil
			},
			append(options)...,
		),
		getCluster: *grpctransport.NewServer(
			endpoints.GetClusterEndpoint,
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
		registerWithRancher: grpctransport.NewServer(
			endpoints.RegisterWithRancherEndpoint,
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

func (s *grpcServer) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	_, rep, err := s.createCluster.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.ClusterResponse), nil
}

func (s *grpcServer) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	_, rep, err := s.addToken.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.AddTokenResponse), nil
}

func (s *grpcServer) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	_, rep, err := s.getToken.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.GetTokenResponse), nil
}

func (s *grpcServer) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	_, rep, err := s.addRoute53Record.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.AddRoute53RecordResponse), nil
}

func (s *grpcServer) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	_, rep, err := s.getClusters.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.GetClustersResponse), nil
}

func (s *grpcServer) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	_, rep, err := s.getCluster.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.ClusterSpec), nil
}

func (s *grpcServer) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	_, rep, err := s.clusterStatus.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.ClusterStatusResponse), nil
}

func (s *grpcServer) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	_, rep, err := s.addNode.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.NodeSpawnResponse), nil
}

func (s *grpcServer) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	_, rep, err := s.deleteCluster.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.ClusterDeleteResponse), nil
}

func (s *grpcServer) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	_, rep, err := s.deleteNode.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.NodeDeleteResponse), nil
}

func (s *grpcServer) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	_, rep, err := s.createVolume.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.CreateVolumeResponse), nil
}

func (s *grpcServer) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	_, rep, err := s.deleteVolume.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.DeleteVolumeResponse), nil
}

func (s *grpcServer) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	_, rep, err := s.createSnapshot.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.CreateSnapshotResponse), nil
}

func (s *grpcServer) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	_, rep, err := s.createSnapshotAndDelete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.CreateSnapshotAndDeleteResponse), nil
}

func (s *grpcServer) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {
	_, rep, err := s.registerWithRancher.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.RancherRegistrationResponse), nil
}

// NewGRPCClient returns an AddService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, logger *zap.SugaredLogger) spawnerservice.ClusterController {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 100))

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
			"proto.SpawnerService",
			"CreateCluster",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.ClusterResponse{},
			append(options)...,
		).Endpoint()
		createClusterEndpoint = limiter(createClusterEndpoint)
		createClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateCluster",
			Timeout: 30 * time.Second,
		}))(createClusterEndpoint)
	}

	var getClustersEndpoint endpoint.Endpoint
	{
		getClustersEndpoint = grpctransport.NewClient(
			conn,
			"proto.SpawnerService",
			"GetClusters",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.GetClustersResponse{},
			append(options)...,
		).Endpoint()
		getClustersEndpoint = limiter(getClustersEndpoint)
		getClustersEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetClusters",
			Timeout: 30 * time.Second,
		}))(getClustersEndpoint)
	}

	var getClusterEndpoint endpoint.Endpoint
	{
		getClusterEndpoint = grpctransport.NewClient(
			conn,
			"proto.SpawnerService",
			"GetCluster",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.ClusterSpec{},
			append(options)...,
		).Endpoint()
		getClusterEndpoint = limiter(getClusterEndpoint)
		getClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetCluster",
			Timeout: 30 * time.Second,
		}))(getClusterEndpoint)
	}

	var addTokenEndpoint endpoint.Endpoint
	{
		addTokenEndpoint = grpctransport.NewClient(
			conn,
			"proto.SpawnerService",
			"AddToken",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.AddTokenResponse{},
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
			"proto.SpawnerService",
			"GetToken",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.GetTokenResponse{},
			append(options)...,
		).Endpoint()
		getTokenEndpoint = limiter(getTokenEndpoint)
		getTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetToken",
			Timeout: 30 * time.Second,
		}))(getTokenEndpoint)
	}

	var addRoute53RecordEndpoint endpoint.Endpoint
	{
		addRoute53RecordEndpoint = grpctransport.NewClient(
			conn,
			"proto.SpawnerService",
			"AddRoute53Record",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.AddRoute53RecordResponse{},
			append(options)...,
		).Endpoint()
		addRoute53RecordEndpoint = limiter(addRoute53RecordEndpoint)
		addRoute53RecordEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "AddRoute53Record",
			Timeout: 30 * time.Second,
		}))(addRoute53RecordEndpoint)
	}

	var clusterStatusEndpoint endpoint.Endpoint
	{
		clusterStatusEndpoint = grpctransport.NewClient(
			conn,
			"proto.SpawnerService",
			"ClusterStatus",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.ClusterStatusResponse{},
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
			"proto.SpawnerService",
			"AddNode",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.NodeSpawnResponse{},
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
			"proto.SpawnerService",
			"DeleteCluster",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.ClusterDeleteResponse{},
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
			"proto.SpawnerService",
			"DeleteNode",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.NodeDeleteResponse{},
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
			"proto.SpawnerService",
			"CreateVolume",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.CreateVolumeResponse{},
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
			"proto.SpawnerService",
			"DeleteVolume",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.DeleteVolumeResponse{},
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
			"proto.SpawnerService",
			"CreateSnapshot",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.CreateSnapshotResponse{},
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
			"proto.SpawnerService",
			"CreateSnapshotAndDelete",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.CreateSnapshotAndDeleteResponse{},
			append(options)...,
		).Endpoint()
		createSnapshotAndDeleteEndpoint = limiter(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateSnapshotAndDelete",
			Timeout: 30 * time.Second,
		}))(createSnapshotAndDeleteEndpoint)
	}

	var registerWithRancherEndpoint endpoint.Endpoint
	{
		registerWithRancherEndpoint = grpctransport.NewClient(
			conn,
			"proto.SpawnerService",
			"RegisterWithRancher",
			func(_ context.Context, grpcReq interface{}) (interface{}, error) {
				return grpcReq, nil
			},
			func(_ context.Context, grpcResp interface{}) (interface{}, error) {
				return grpcResp, nil
			},
			proto.RancherRegistrationResponse{},
			append(options)...,
		).Endpoint()
		registerWithRancherEndpoint = limiter(registerWithRancherEndpoint)
		registerWithRancherEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "RegisterWithRancher",
			Timeout: 30 * time.Second,
		}))(registerWithRancherEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return spwnendpoint.Set{
		CreateClusterEndpoint:           createClusterEndpoint,
		AddTokenEndpoint:                addTokenEndpoint,
		GetTokenEndpoint:                getTokenEndpoint,
		AddRoute53RecordEndpoint:        addRoute53RecordEndpoint,
		GetClustersEndpoint:             getClustersEndpoint,
		GetClusterEndpoint:              getClusterEndpoint,
		CusterStatusEndpoint:            clusterStatusEndpoint,
		AddNodeEndpoint:                 addNodeEndpoint,
		DeleteClusterEndpoint:           deleteClusterEndpoint,
		DeleteNodeEndpoint:              deleteNodeEndpoint,
		CreateVolumeEndpoint:            createVolumeEndpoint,
		DeleteVolumeEndpoint:            deleteVolumeEndpoint,
		CreateSnapshotEndpoint:          createSnapshotEndpoint,
		CreateSnapshotAndDeleteEndpoint: createSnapshotAndDeleteEndpoint,
		RegisterWithRancherEndpoint:     registerWithRancherEndpoint,
	}
}

// decodeGRPCSumRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func decodeGRPCClusterRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	// req := grpcReq.(*proto.ClusterRequest)
	// return req, nil
	return grpcReq, nil
}

// encodeGRPCSumResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain sum response to a gRPC sum reply. Primarily useful in a server.
func encodeGRPCClusterResponse(_ context.Context, response interface{}) (interface{}, error) {
	// resp := response.(*proto.ClusterResponse)
	// return &resp, nil
	return response, nil
}
