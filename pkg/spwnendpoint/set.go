package spwnendpoint

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/sony/gobreaker"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Set collects all of the endpoints that compose an add service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type Set struct {
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
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc spawnerservice.ClusterController, logger *zap.SugaredLogger, duration metrics.Histogram) Set {
	var createClusterEndpoint endpoint.Endpoint
	{
		createClusterEndpoint = MakeCreateClusterEndpoint(svc)
		// CreateCluster is limited to 1 request per second with burst of 1 request.
		// Note, rate is defined as a time interval between requests.
		createClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createClusterEndpoint)
		createClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createClusterEndpoint)
		createClusterEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "CreateCluster"))(createClusterEndpoint)
		createClusterEndpoint = InstrumentingMiddleware(duration.With("method", "CreateCluster"))(createClusterEndpoint)
	}

	var getClustersEndpoint endpoint.Endpoint
	{
		getClustersEndpoint = MakeGetClustersEndpoint(svc)
		// GetClusters is limited to 1 request per second with burst of 1 request.
		// Note, rate is defined as a time interval between requests.
		getClustersEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getClustersEndpoint)
		getClustersEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getClustersEndpoint)
		getClustersEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "GetClusters"))(getClustersEndpoint)
		getClustersEndpoint = InstrumentingMiddleware(duration.With("method", "GetClusters"))(getClustersEndpoint)
	}

	var getClusterEndpoint endpoint.Endpoint
	{
		getClusterEndpoint = MakeGetClusterEndpoint(svc)
		getClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getClusterEndpoint)
		getClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getClusterEndpoint)
		getClusterEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "GetCluster"))(getClusterEndpoint)
		getClusterEndpoint = InstrumentingMiddleware(duration.With("method", "GetCluster"))(getClusterEndpoint)
	}

	var addTokenEndpoint endpoint.Endpoint
	{
		addTokenEndpoint = MakeAddTokenEndpoint(svc)
		addTokenEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(addTokenEndpoint)
		addTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addTokenEndpoint)
		addTokenEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "AddToken"))(addTokenEndpoint)
		addTokenEndpoint = InstrumentingMiddleware(duration.With("method", "AddToken"))(addTokenEndpoint)
	}

	var getTokenEndpoint endpoint.Endpoint
	{
		getTokenEndpoint = MakeGetTokenEndpoint(svc)
		getTokenEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getTokenEndpoint)
		getTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getTokenEndpoint)
		getTokenEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "GetToken"))(getTokenEndpoint)
		getTokenEndpoint = InstrumentingMiddleware(duration.With("method", "GetToken"))(getTokenEndpoint)
	}

	var addRoute53RecordEndpoint endpoint.Endpoint
	{
		addRoute53RecordEndpoint = MakeAddRoute53RecordEndpoint(svc)
		addRoute53RecordEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(addRoute53RecordEndpoint)
		addRoute53RecordEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addRoute53RecordEndpoint)
		addRoute53RecordEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "AddRoute53Record"))(addRoute53RecordEndpoint)
		addRoute53RecordEndpoint = InstrumentingMiddleware(duration.With("method", "AddRoute53Record"))(addRoute53RecordEndpoint)
	}

	var clusterStatusEndpoint endpoint.Endpoint
	{
		clusterStatusEndpoint = MakeCusterStatusEndpoint(svc)
		clusterStatusEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(clusterStatusEndpoint)
		clusterStatusEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(clusterStatusEndpoint)
		clusterStatusEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "ClusterStatus"))(clusterStatusEndpoint)
		clusterStatusEndpoint = InstrumentingMiddleware(duration.With("method", "ClusterStatus"))(clusterStatusEndpoint)
	}

	var addNodeEndpoint endpoint.Endpoint
	{
		addNodeEndpoint = MakeAddNodeEndpoint(svc)
		addNodeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(addNodeEndpoint)
		addNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addNodeEndpoint)
		addNodeEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "AddNode"))(addNodeEndpoint)
		addNodeEndpoint = InstrumentingMiddleware(duration.With("method", "AddNode"))(addNodeEndpoint)
	}

	var deleteClusterEndpoint endpoint.Endpoint
	{
		deleteClusterEndpoint = MakeClusterDeleteEndpoint(svc)
		deleteClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(deleteClusterEndpoint)
		deleteClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteClusterEndpoint)
		deleteClusterEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "DeleteCluster"))(deleteClusterEndpoint)
		deleteClusterEndpoint = InstrumentingMiddleware(duration.With("method", "DeleteCluster"))(deleteClusterEndpoint)
	}

	var deleteNodeEndpoint endpoint.Endpoint
	{
		deleteNodeEndpoint = MakeNodeDeleteEndpoint(svc)
		deleteNodeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(deleteNodeEndpoint)
		deleteNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteNodeEndpoint)
		deleteNodeEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "DeleteNode"))(deleteNodeEndpoint)
		deleteNodeEndpoint = InstrumentingMiddleware(duration.With("method", "DeleteNode"))(deleteNodeEndpoint)
	}

	var createVolumeEndpoint endpoint.Endpoint
	{
		createVolumeEndpoint = MakeCreateVolumeEndpoint(svc)
		createVolumeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createVolumeEndpoint)
		createVolumeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createVolumeEndpoint)
		createVolumeEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "CreateVolume"))(createVolumeEndpoint)
		createVolumeEndpoint = InstrumentingMiddleware(duration.With("method", "CreateVolume"))(createVolumeEndpoint)
	}

	var deleteVolumeEndpoint endpoint.Endpoint
	{
		deleteVolumeEndpoint = MakeDeleteVolumeEndpoint(svc)
		deleteVolumeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(deleteVolumeEndpoint)
		deleteVolumeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteVolumeEndpoint)
		deleteVolumeEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "DeleteVolume"))(deleteVolumeEndpoint)
		deleteVolumeEndpoint = InstrumentingMiddleware(duration.With("method", "DeleteVolume"))(deleteVolumeEndpoint)
	}

	var createSnapshotEndpoint endpoint.Endpoint
	{
		createSnapshotEndpoint = MakeCreateSnapshotEndpoint(svc)
		createSnapshotEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createSnapshotEndpoint)
		createSnapshotEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createSnapshotEndpoint)
		createSnapshotEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "CreateSnapshot"))(createSnapshotEndpoint)
		createSnapshotEndpoint = InstrumentingMiddleware(duration.With("method", "CreateSnapshot"))(createSnapshotEndpoint)
	}

	var createSnapshotAndDeleteEndpoint endpoint.Endpoint
	{
		createSnapshotAndDeleteEndpoint = MakeCreateSnapshotAndDeleteEndpoint(svc)
		createSnapshotAndDeleteEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "CreateSnapshotAndDelete"))(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = InstrumentingMiddleware(duration.With("method", "CreateSnapshotAndDelete"))(createSnapshotAndDeleteEndpoint)
	}

	var registerWithRancherEndpoint endpoint.Endpoint
	{
		registerWithRancherEndpoint = MakeRegisterWithRancherEndpoint(svc)
		registerWithRancherEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(registerWithRancherEndpoint)
		registerWithRancherEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(registerWithRancherEndpoint)
		registerWithRancherEndpoint = LoggingMiddleware(logger.With("logger", logger, "method", "RegisterWithRancherEndpoint"))(registerWithRancherEndpoint)
		registerWithRancherEndpoint = InstrumentingMiddleware(duration.With("method", "CreateSnapshotAndDelete"))(registerWithRancherEndpoint)
	}

	return Set{
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
func (s Set) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	resp, err := s.AddTokenEndpoint(ctx, req)
	if err != nil {
		return &pb.AddTokenResponse{}, err
	}
	response := resp.(*pb.AddTokenResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeAddTokenEndpoint constructs a AddToken endpoint wrapping the service.
func MakeAddTokenEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.AddTokenRequest)
		resp, err := s.AddToken(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	resp, err := s.GetTokenEndpoint(ctx, req)
	if err != nil {
		return &pb.GetTokenResponse{}, err
	}
	response := resp.(*pb.GetTokenResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeGetTokenEndpoint constructs a GetToken endpoint wrapping the service.
func MakeGetTokenEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.GetTokenRequest)
		resp, err := s.GetToken(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (*pb.AddRoute53RecordResponse, error) {
	resp, err := s.AddRoute53RecordEndpoint(ctx, req)
	if err != nil {
		return &pb.AddRoute53RecordResponse{}, err
	}
	response := resp.(*pb.AddRoute53RecordResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeAddRoute53RecordEndpoint constructs a AddRoute53Record endpoint wrapping the service.
func MakeAddRoute53RecordEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.AddRoute53RecordRequest)
		resp, err := s.AddRoute53Record(ctx, req)
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

func (s Set) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	resp, err := s.CreateVolumeEndpoint(ctx, req)
	if err != nil {
		return &pb.CreateVolumeResponse{}, err
	}
	response := resp.(*pb.CreateVolumeResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeCreateVolumeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.CreateVolumeRequest)
		resp, err := s.CreateVolume(ctx, req)
		return resp, err
	}
}

func (s Set) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	resp, err := s.DeleteVolumeEndpoint(ctx, req)
	if err != nil {
		return &pb.DeleteVolumeResponse{}, err
	}
	response := resp.(*pb.DeleteVolumeResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeDeleteVolumeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.DeleteVolumeRequest)
		resp, err := s.DeleteVolume(ctx, req)
		return resp, err
	}
}

func (s Set) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	resp, err := s.CreateSnapshotEndpoint(ctx, req)
	if err != nil {
		return &pb.CreateSnapshotResponse{}, err
	}
	response := resp.(*pb.CreateSnapshotResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeCreateSnapshotEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.CreateSnapshotRequest)
		resp, err := s.CreateSnapshot(ctx, req)
		return resp, err
	}
}

func (s Set) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	resp, err := s.CreateSnapshotAndDeleteEndpoint(ctx, req)
	if err != nil {
		return &pb.CreateSnapshotAndDeleteResponse{}, err
	}
	response := resp.(*pb.CreateSnapshotAndDeleteResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeCreateSnapshotAndDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.CreateSnapshotAndDeleteRequest)
		resp, err := s.CreateSnapshotAndDelete(ctx, req)
		return resp, err
	}
}

func (s Set) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (*pb.GetClustersResponse, error) {
	resp, err := s.GetClustersEndpoint(ctx, req)
	if err != nil {
		return &pb.GetClustersResponse{}, err
	}
	response := resp.(*pb.GetClustersResponse)
	return response, fmt.Errorf("")
}

func MakeGetClustersEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.GetClustersRequest)
		resp, err := s.GetClusters(ctx, req)
		return resp, err
	}
}

func (s Set) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (*pb.ClusterSpec, error) {

	resp, err := s.GetClusterEndpoint(ctx, req)
	if err != nil {
		return &pb.ClusterSpec{}, err
	}
	response := resp.(*pb.ClusterSpec)
	return response, fmt.Errorf("")
}

func MakeGetClusterEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.GetClusterRequest)
		resp, err := s.GetCluster(ctx, req)
		return resp, err
	}
}

func (s Set) RegisterWithRancher(ctx context.Context, req *pb.RancherRegistrationRequest) (*pb.RancherRegistrationResponse, error) {
	fmt.Println(" register with rancher")
	resp, err := s.RegisterWithRancherEndpoint(ctx, req)
	if err != nil {
		return &pb.RancherRegistrationResponse{}, err
	}
	response := resp.(*pb.RancherRegistrationResponse)
	return response, fmt.Errorf("")
}

func MakeRegisterWithRancherEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.RancherRegistrationRequest)
		resp, err := s.RegisterWithRancher(ctx, req)
		return resp, err
	}
}
