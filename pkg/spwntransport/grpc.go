package spwntransport

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"

	kitzap "github.com/go-kit/kit/log/zap"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"

	"go.uber.org/zap"
)

type grpcServer struct {
	createCluster           grpctransport.Handler
	addToken                grpctransport.Handler
	getToken                grpctransport.Handler
	addRoute53Record        grpctransport.Handler
	getClusters             grpctransport.Handler
	getCluster              grpctransport.Handler
	clusterStatus           grpctransport.Handler
	addNode                 grpctransport.Handler
	deleteCluster           grpctransport.Handler
	deleteNode              grpctransport.Handler
	createVolume            grpctransport.Handler
	deleteVolume            grpctransport.Handler
	createSnapshot          grpctransport.Handler
	createSnapshotAndDelete grpctransport.Handler
	registerWithRancher     grpctransport.Handler
	getWorkspaceCost        grpctransport.Handler
	readCredential          grpctransport.Handler
	writeCredential         grpctransport.Handler

	proto.UnimplementedSpawnerServiceServer
}

func newServer(endpoint endpoint.Endpoint, options []grpctransport.ServerOption) *grpctransport.Server {
	return grpctransport.NewServer(
		endpoint,
		func(_ context.Context, grpcReq interface{}) (interface{}, error) {
			return grpcReq, nil
		},
		func(_ context.Context, response interface{}) (interface{}, error) {
			return response, nil
		},
		options...,
	)
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints spwnendpoint.Set, logger *zap.SugaredLogger) proto.SpawnerServiceServer {
	kitZapLogger := kitzap.NewZapSugarLogger(logger.Desugar(), zap.InfoLevel)
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(kitZapLogger)),
	}

	return &grpcServer{
		createCluster:           newServer(endpoints.CreateClusterEndpoint, options),
		addToken:                newServer(endpoints.AddTokenEndpoint, options),
		getToken:                newServer(endpoints.GetTokenEndpoint, options),
		addRoute53Record:        newServer(endpoints.AddRoute53RecordEndpoint, options),
		getClusters:             newServer(endpoints.GetClustersEndpoint, options),
		getCluster:              newServer(endpoints.GetClusterEndpoint, options),
		clusterStatus:           newServer(endpoints.GetClusterEndpoint, options),
		addNode:                 newServer(endpoints.AddNodeEndpoint, options),
		deleteCluster:           newServer(endpoints.DeleteClusterEndpoint, options),
		deleteNode:              newServer(endpoints.DeleteNodeEndpoint, options),
		createVolume:            newServer(endpoints.CreateVolumeEndpoint, options),
		deleteVolume:            newServer(endpoints.DeleteVolumeEndpoint, options),
		createSnapshot:          newServer(endpoints.CreateSnapshotEndpoint, options),
		createSnapshotAndDelete: newServer(endpoints.CreateSnapshotAndDeleteEndpoint, options),
		registerWithRancher:     newServer(endpoints.RegisterWithRancherEndpoint, options),
		getWorkspaceCost:        newServer(endpoints.GetWorkspaceCostEndpoint, options),
		writeCredential:         newServer(endpoints.WriteCredentialEndpoint, options),
		readCredential:          newServer(endpoints.ReadCredentialEndpoint, options),
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

func (s *grpcServer) GetWorkspaceCost(ctx context.Context, req *proto.GetWorkspaceCostRequest) (*proto.GetWorkspaceCostResponse, error) {
	_, rep, err := s.getWorkspaceCost.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.GetWorkspaceCostResponse), nil
}

func (s *grpcServer) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {
	_, rep, err := s.writeCredential.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.WriteCredentialResponse), nil
}

func (s *grpcServer) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {
	_, rep, err := s.readCredential.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.ReadCredentialResponse), nil
}
