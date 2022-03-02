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
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
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
	GetWorkspaceCostEndpoint        endpoint.Endpoint
	ReadCredentialEndpoint          endpoint.Endpoint
	WriteCredentialEndpoint         endpoint.Endpoint
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
		createClusterEndpoint = InstrumentingMiddleware(duration.With("method", "CreateCluster"))(createClusterEndpoint)
	}

	var getClustersEndpoint endpoint.Endpoint
	{
		getClustersEndpoint = MakeGetClustersEndpoint(svc)
		// GetClusters is limited to 1 request per second with burst of 1 request.
		// Note, rate is defined as a time interval between requests.
		getClustersEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getClustersEndpoint)
		getClustersEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getClustersEndpoint)
		getClustersEndpoint = InstrumentingMiddleware(duration.With("method", "GetClusters"))(getClustersEndpoint)
	}

	var getClusterEndpoint endpoint.Endpoint
	{
		getClusterEndpoint = MakeGetClusterEndpoint(svc)
		getClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getClusterEndpoint)
		getClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getClusterEndpoint)
		getClusterEndpoint = InstrumentingMiddleware(duration.With("method", "GetCluster"))(getClusterEndpoint)
	}

	var addTokenEndpoint endpoint.Endpoint
	{
		addTokenEndpoint = MakeAddTokenEndpoint(svc)
		addTokenEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(addTokenEndpoint)
		addTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addTokenEndpoint)
		addTokenEndpoint = InstrumentingMiddleware(duration.With("method", "AddToken"))(addTokenEndpoint)
	}

	var getTokenEndpoint endpoint.Endpoint
	{
		getTokenEndpoint = MakeGetTokenEndpoint(svc)
		getTokenEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getTokenEndpoint)
		getTokenEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getTokenEndpoint)
		getTokenEndpoint = InstrumentingMiddleware(duration.With("method", "GetToken"))(getTokenEndpoint)
	}

	var addRoute53RecordEndpoint endpoint.Endpoint
	{
		addRoute53RecordEndpoint = MakeAddRoute53RecordEndpoint(svc)
		addRoute53RecordEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(addRoute53RecordEndpoint)
		addRoute53RecordEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addRoute53RecordEndpoint)
		addRoute53RecordEndpoint = InstrumentingMiddleware(duration.With("method", "AddRoute53Record"))(addRoute53RecordEndpoint)
	}

	var clusterStatusEndpoint endpoint.Endpoint
	{
		clusterStatusEndpoint = MakeCusterStatusEndpoint(svc)
		clusterStatusEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(clusterStatusEndpoint)
		clusterStatusEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(clusterStatusEndpoint)
		clusterStatusEndpoint = InstrumentingMiddleware(duration.With("method", "ClusterStatus"))(clusterStatusEndpoint)
	}

	var addNodeEndpoint endpoint.Endpoint
	{
		addNodeEndpoint = MakeAddNodeEndpoint(svc)
		addNodeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(addNodeEndpoint)
		addNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(addNodeEndpoint)
		addNodeEndpoint = InstrumentingMiddleware(duration.With("method", "AddNode"))(addNodeEndpoint)
	}

	var deleteClusterEndpoint endpoint.Endpoint
	{
		deleteClusterEndpoint = MakeClusterDeleteEndpoint(svc)
		deleteClusterEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(deleteClusterEndpoint)
		deleteClusterEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteClusterEndpoint)
		deleteClusterEndpoint = InstrumentingMiddleware(duration.With("method", "DeleteCluster"))(deleteClusterEndpoint)
	}

	var deleteNodeEndpoint endpoint.Endpoint
	{
		deleteNodeEndpoint = MakeNodeDeleteEndpoint(svc)
		deleteNodeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(deleteNodeEndpoint)
		deleteNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteNodeEndpoint)
		deleteNodeEndpoint = InstrumentingMiddleware(duration.With("method", "DeleteNode"))(deleteNodeEndpoint)
	}

	var createVolumeEndpoint endpoint.Endpoint
	{
		createVolumeEndpoint = MakeCreateVolumeEndpoint(svc)
		createVolumeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createVolumeEndpoint)
		createVolumeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createVolumeEndpoint)
		createVolumeEndpoint = InstrumentingMiddleware(duration.With("method", "CreateVolume"))(createVolumeEndpoint)
	}

	var deleteVolumeEndpoint endpoint.Endpoint
	{
		deleteVolumeEndpoint = MakeDeleteVolumeEndpoint(svc)
		deleteVolumeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(deleteVolumeEndpoint)
		deleteVolumeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(deleteVolumeEndpoint)
		deleteVolumeEndpoint = InstrumentingMiddleware(duration.With("method", "DeleteVolume"))(deleteVolumeEndpoint)
	}

	var createSnapshotEndpoint endpoint.Endpoint
	{
		createSnapshotEndpoint = MakeCreateSnapshotEndpoint(svc)
		createSnapshotEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createSnapshotEndpoint)
		createSnapshotEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createSnapshotEndpoint)
		createSnapshotEndpoint = InstrumentingMiddleware(duration.With("method", "CreateSnapshot"))(createSnapshotEndpoint)
	}

	var createSnapshotAndDeleteEndpoint endpoint.Endpoint
	{
		createSnapshotAndDeleteEndpoint = MakeCreateSnapshotAndDeleteEndpoint(svc)
		createSnapshotAndDeleteEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createSnapshotAndDeleteEndpoint)
		createSnapshotAndDeleteEndpoint = InstrumentingMiddleware(duration.With("method", "CreateSnapshotAndDelete"))(createSnapshotAndDeleteEndpoint)
	}

	var registerWithRancherEndpoint endpoint.Endpoint
	{
		registerWithRancherEndpoint = MakeRegisterWithRancherEndpoint(svc)
		registerWithRancherEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(registerWithRancherEndpoint)
		registerWithRancherEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(registerWithRancherEndpoint)
		registerWithRancherEndpoint = InstrumentingMiddleware(duration.With("method", "CreateSnapshotAndDelete"))(registerWithRancherEndpoint)
	}

	var getWorkspaceCostEndpoint endpoint.Endpoint
	{
		getWorkspaceCostEndpoint = MakeGetWorkspaceCostEndpoint(svc)
		getWorkspaceCostEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(getWorkspaceCostEndpoint)
		getWorkspaceCostEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getWorkspaceCostEndpoint)
		getWorkspaceCostEndpoint = InstrumentingMiddleware(duration.With("method", "GetWorkspaceCost"))(getWorkspaceCostEndpoint)
	}

	var readCredentialsEndpoint endpoint.Endpoint
	{
		readCredentialsEndpoint = MakeReadCredentialsEndpoint(svc)
		readCredentialsEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(readCredentialsEndpoint)
		readCredentialsEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(readCredentialsEndpoint)
		readCredentialsEndpoint = InstrumentingMiddleware(duration.With("method", "ReadCredential"))(readCredentialsEndpoint)
	}

	var writeCredentialsEndpoint endpoint.Endpoint
	{
		writeCredentialsEndpoint = MakeWriteCredentialsEndpoint(svc)
		writeCredentialsEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second/10), 1))(writeCredentialsEndpoint)
		writeCredentialsEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(writeCredentialsEndpoint)
		writeCredentialsEndpoint = InstrumentingMiddleware(duration.With("method", "WriteCredential"))(writeCredentialsEndpoint)
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
		GetWorkspaceCostEndpoint:        getWorkspaceCostEndpoint,
		ReadCredentialEndpoint:          readCredentialsEndpoint,
		WriteCredentialEndpoint:         writeCredentialsEndpoint,
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	resp, err := s.CreateClusterEndpoint(ctx, req)
	if err != nil {
		return &proto.ClusterResponse{}, err
	}
	response := resp.(*proto.ClusterResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeCreateClusterEndpoint constructs a CreateCluster endpoint wrapping the service.
func MakeCreateClusterEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ClusterRequest)
		resp, err := s.CreateCluster(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	resp, err := s.AddTokenEndpoint(ctx, req)
	if err != nil {
		return &proto.AddTokenResponse{}, err
	}
	response := resp.(*proto.AddTokenResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeAddTokenEndpoint constructs a AddToken endpoint wrapping the service.
func MakeAddTokenEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.AddTokenRequest)
		resp, err := s.AddToken(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	resp, err := s.GetTokenEndpoint(ctx, req)
	if err != nil {
		return &proto.GetTokenResponse{}, err
	}
	response := resp.(*proto.GetTokenResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeGetTokenEndpoint constructs a GetToken endpoint wrapping the service.
func MakeGetTokenEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetTokenRequest)
		resp, err := s.GetToken(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	resp, err := s.AddRoute53RecordEndpoint(ctx, req)
	if err != nil {
		return &proto.AddRoute53RecordResponse{}, err
	}
	response := resp.(*proto.AddRoute53RecordResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeAddRoute53RecordEndpoint constructs a AddRoute53Record endpoint wrapping the service.
func MakeAddRoute53RecordEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.AddRoute53RecordRequest)
		resp, err := s.AddRoute53Record(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	resp, err := s.CusterStatusEndpoint(ctx, req)
	if err != nil {
		return &proto.ClusterStatusResponse{}, err
	}
	response := resp.(*proto.ClusterStatusResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeCusterStatusEndpoint constructs a ClusterStatus endpoint wrapping the service.
func MakeCusterStatusEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ClusterStatusRequest)
		resp, err := s.ClusterStatus(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	resp, err := s.AddNodeEndpoint(ctx, req)
	if err != nil {
		return &proto.NodeSpawnResponse{}, err
	}
	response := resp.(*proto.NodeSpawnResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeAddNodeEndpoint constructs a AddNode endpoint wrapping the service.
func MakeAddNodeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.NodeSpawnRequest)
		resp, err := s.AddNode(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	resp, err := s.DeleteClusterEndpoint(ctx, req)
	if err != nil {
		return &proto.ClusterDeleteResponse{}, err
	}
	response := resp.(*proto.ClusterDeleteResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeClusterDeleteEndpointt constructs a ClusterStatus endpoint wrapping the service.
func MakeClusterDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ClusterDeleteRequest)
		resp, err := s.DeleteCluster(ctx, req)
		return resp, err
	}
}

// Implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	resp, err := s.DeleteNodeEndpoint(ctx, req)
	if err != nil {
		return &proto.NodeDeleteResponse{}, err
	}
	response := resp.(*proto.NodeDeleteResponse)
	return response, fmt.Errorf(response.Error)
}

// MakeClusterDeleteEndpointt constructs a ClusterStatus endpoint wrapping the service.
func MakeNodeDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.NodeDeleteRequest)
		resp, err := s.DeleteNode(ctx, req)
		return resp, err
	}
}

func (s Set) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	resp, err := s.CreateVolumeEndpoint(ctx, req)
	if err != nil {
		return &proto.CreateVolumeResponse{}, err
	}
	response := resp.(*proto.CreateVolumeResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeCreateVolumeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.CreateVolumeRequest)
		resp, err := s.CreateVolume(ctx, req)
		return resp, err
	}
}

func (s Set) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	resp, err := s.DeleteVolumeEndpoint(ctx, req)
	if err != nil {
		return &proto.DeleteVolumeResponse{}, err
	}
	response := resp.(*proto.DeleteVolumeResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeDeleteVolumeEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.DeleteVolumeRequest)
		resp, err := s.DeleteVolume(ctx, req)
		return resp, err
	}
}

func (s Set) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	resp, err := s.CreateSnapshotEndpoint(ctx, req)
	if err != nil {
		return &proto.CreateSnapshotResponse{}, err
	}
	response := resp.(*proto.CreateSnapshotResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeCreateSnapshotEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.CreateSnapshotRequest)
		resp, err := s.CreateSnapshot(ctx, req)
		return resp, err
	}
}

func (s Set) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	resp, err := s.CreateSnapshotAndDeleteEndpoint(ctx, req)
	if err != nil {
		return &proto.CreateSnapshotAndDeleteResponse{}, err
	}
	response := resp.(*proto.CreateSnapshotAndDeleteResponse)
	return response, fmt.Errorf(response.Error)
}

func MakeCreateSnapshotAndDeleteEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.CreateSnapshotAndDeleteRequest)
		resp, err := s.CreateSnapshotAndDelete(ctx, req)
		return resp, err
	}
}

func (s Set) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	resp, err := s.GetClustersEndpoint(ctx, req)
	if err != nil {
		return &proto.GetClustersResponse{}, err
	}
	response := resp.(*proto.GetClustersResponse)
	return response, fmt.Errorf("")
}

func MakeGetClustersEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetClustersRequest)
		resp, err := s.GetClusters(ctx, req)
		return resp, err
	}
}

func (s Set) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	resp, err := s.GetClusterEndpoint(ctx, req)
	if err != nil {
		return &proto.ClusterSpec{}, err
	}
	response := resp.(*proto.ClusterSpec)
	return response, fmt.Errorf("")
}

func MakeGetClusterEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetClusterRequest)
		resp, err := s.GetCluster(ctx, req)
		return resp, err
	}
}

func (s Set) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {
	resp, err := s.RegisterWithRancherEndpoint(ctx, req)
	if err != nil {
		return &proto.RancherRegistrationResponse{}, err
	}
	response := resp.(*proto.RancherRegistrationResponse)
	return response, fmt.Errorf("")
}

func MakeRegisterWithRancherEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.RancherRegistrationRequest)
		resp, err := s.RegisterWithRancher(ctx, req)
		return resp, err
	}
}

func MakeGetWorkspaceCostEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.GetWorkspaceCostRequest)
		resp, err := s.GetWorkspaceCost(ctx, req)
		return resp, err
	}
}

func (s Set) GetWorkspaceCost(ctx context.Context, req *proto.GetWorkspaceCostRequest) (*proto.GetWorkspaceCostResponse, error) {

	resp, err := s.GetWorkspaceCostEndpoint(ctx, req)
	if err != nil {
		return &proto.GetWorkspaceCostResponse{}, err
	}
	response := resp.(*proto.GetWorkspaceCostResponse)
	return response, nil
}

func MakeWriteCredentialsEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.WriteCredentialRequest)
		resp, err := s.WriteCredential(ctx, req)
		return resp, err
	}
}

func (s Set) WriteCredential(ctx context.Context, req *proto.WriteCredentialRequest) (*proto.WriteCredentialResponse, error) {

	resp, err := s.WriteCredentialEndpoint(ctx, req)
	if err != nil {
		return &proto.WriteCredentialResponse{}, err
	}
	response := resp.(*proto.WriteCredentialResponse)
	return response, nil
}

func MakeReadCredentialsEndpoint(s spawnerservice.ClusterController) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.ReadCredentialRequest)
		resp, err := s.ReadCredential(ctx, req)
		return resp, err
	}
}

func (s Set) ReadCredential(ctx context.Context, req *proto.ReadCredentialRequest) (*proto.ReadCredentialResponse, error) {

	resp, err := s.ReadCredentialEndpoint(ctx, req)
	if err != nil {
		return &proto.ReadCredentialResponse{}, err
	}
	response := resp.(*proto.ReadCredentialResponse)
	return response, nil
}
