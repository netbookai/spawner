package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
)

// SpawnerEndpoints collects all of the endpoints that compose an add service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type SpawnerEndpoints struct {
	Echo                            endpoint.Endpoint
	HealthCheck                     endpoint.Endpoint
	CreateClusterEndpoint           endpoint.Endpoint
	AddTokenEndpoint                endpoint.Endpoint
	AddRoute53RecordEndpoint        endpoint.Endpoint
	GetTokenEndpoint                endpoint.Endpoint
	GetClustersEndpoint             endpoint.Endpoint
	GetClusterEndpoint              endpoint.Endpoint
	CusterStatusEndpoint            endpoint.Endpoint
	AddNodeEndpoint                 endpoint.Endpoint
	DeleteClusterEndpoint           endpoint.Endpoint
	DeleteNodeEndpoint              endpoint.Endpoint
	CreateVolumeEndpoint            endpoint.Endpoint
	DeleteVolumeEndpoint            endpoint.Endpoint
	CreateSnapshotEndpoint          endpoint.Endpoint
	CreateSnapshotAndDeleteEndpoint endpoint.Endpoint
	RegisterWithRancherEndpoint     endpoint.Endpoint
	GetWorkspaceCostEndpoint        endpoint.Endpoint
	ReadCredentialEndpoint          endpoint.Endpoint
	WriteCredentialEndpoint         endpoint.Endpoint
}

// New returns a SpawnerEndpoints that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc spawnerservice.ClusterController, logger *zap.SugaredLogger) *SpawnerEndpoints {

	echoEndpoint := makeEchoEndpoint()
	healthCheckEndpoint := makeHelathCheckEndpoint()
	createClusterEndpoint := makeCreateClusterEndpoint(svc)
	getClustersEndpoint := makeGetClustersEndpoint(svc)
	getClusterEndpoint := makeGetClusterEndpoint(svc)
	addTokenEndpoint := makeAddTokenEndpoint(svc)
	getTokenEndpoint := makeGetTokenEndpoint(svc)
	addRoute53RecordEndpoint := makeAddRoute53RecordEndpoint(svc)
	clusterStatusEndpoint := makeCusterStatusEndpoint(svc)
	addNodeEndpoint := makeAddNodeEndpoint(svc)
	deleteClusterEndpoint := makeClusterDeleteEndpoint(svc)
	deleteNodeEndpoint := makeNodeDeleteEndpoint(svc)
	createVolumeEndpoint := makeCreateVolumeEndpoint(svc)
	deleteVolumeEndpoint := makeDeleteVolumeEndpoint(svc)
	createSnapshotEndpoint := makeCreateSnapshotEndpoint(svc)
	createSnapshotAndDeleteEndpoint := makeCreateSnapshotAndDeleteEndpoint(svc)
	registerWithRancherEndpoint := makeRegisterWithRancherEndpoint(svc)
	getWorkspaceCostEndpoint := makeGetWorkspaceCostEndpoint(svc)
	readCredentialsEndpoint := makeReadCredentialsEndpoint(svc)
	writeCredentialsEndpoint := makeWriteCredentialsEndpoint(svc)

	return &SpawnerEndpoints{
		Echo:                            echoEndpoint,
		HealthCheck:                     healthCheckEndpoint,
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
		GetWorkspaceCostEndpoint:        getWorkspaceCostEndpoint,
		ReadCredentialEndpoint:          readCredentialsEndpoint,
		WriteCredentialEndpoint:         writeCredentialsEndpoint,
	}
}

func makeHelathCheckEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return &proto.Empty{}, nil
	}
}

func makeEchoEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.EchoRequest)
		return &proto.EchoResponse{Msg: req.GetMsg()}, nil
	}
}

// makeCreateClusterEndpoint constructs a CreateCluster endpoint wrapping the service.
func makeCreateClusterEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ClusterRequest)
		resp, err := s.CreateCluster(ctx, req)
		return resp, err
	}
}

// makeAddTokenEndpoint constructs a AddToken endpoint wrapping the service.
func makeAddTokenEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.AddTokenRequest)
		resp, err := s.AddToken(ctx, req)
		return resp, err
	}
}

// makeGetTokenEndpoint constructs a GetToken endpoint wrapping the service.
func makeGetTokenEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetTokenRequest)
		resp, err := s.GetToken(ctx, req)
		return resp, err
	}
}

// makeAddRoute53RecordEndpoint constructs a AddRoute53Record endpoint wrapping the service.
func makeAddRoute53RecordEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.AddRoute53RecordRequest)
		resp, err := s.AddRoute53Record(ctx, req)
		return resp, err
	}
}

// makeCusterStatusEndpoint constructs a ClusterStatus endpoint wrapping the service.
func makeCusterStatusEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ClusterStatusRequest)
		resp, err := s.ClusterStatus(ctx, req)
		return resp, err
	}
}

// makeAddNodeEndpoint constructs a AddNode endpoint wrapping the service.
func makeAddNodeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.NodeSpawnRequest)
		resp, err := s.AddNode(ctx, req)
		return resp, err
	}
}

// makeClusterDeleteEndpointt constructs a ClusterStatus endpoint wrapping the service.
func makeClusterDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ClusterDeleteRequest)
		resp, err := s.DeleteCluster(ctx, req)
		return resp, err
	}
}

// makeNodeDeleteEndpoint constructs a ClusterStatus endpoint wrapping the service.
func makeNodeDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.NodeDeleteRequest)
		resp, err := s.DeleteNode(ctx, req)
		return resp, err
	}
}

func makeCreateVolumeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.CreateVolumeRequest)
		resp, err := s.CreateVolume(ctx, req)
		return resp, err
	}
}

func makeDeleteVolumeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.DeleteVolumeRequest)
		resp, err := s.DeleteVolume(ctx, req)
		return resp, err
	}
}

func makeCreateSnapshotEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.CreateSnapshotRequest)
		resp, err := s.CreateSnapshot(ctx, req)
		return resp, err
	}
}

func makeCreateSnapshotAndDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.CreateSnapshotAndDeleteRequest)
		resp, err := s.CreateSnapshotAndDelete(ctx, req)
		return resp, err
	}
}

func makeGetClustersEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetClustersRequest)
		resp, err := s.GetClusters(ctx, req)
		return resp, err
	}
}

func makeGetClusterEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetClusterRequest)
		resp, err := s.GetCluster(ctx, req)
		return resp, err
	}
}

func makeRegisterWithRancherEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.RancherRegistrationRequest)
		resp, err := s.RegisterWithRancher(ctx, req)
		return resp, err
	}
}

func makeGetWorkspaceCostEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetWorkspaceCostRequest)
		resp, err := s.GetWorkspaceCost(ctx, req)
		return resp, err
	}
}

func makeWriteCredentialsEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.WriteCredentialRequest)
		resp, err := s.WriteCredential(ctx, req)
		return resp, err
	}
}

func makeReadCredentialsEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ReadCredentialRequest)
		resp, err := s.ReadCredential(ctx, req)
		return resp, err
	}
}
