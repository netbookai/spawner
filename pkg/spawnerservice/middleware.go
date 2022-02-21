package spawnerservice

import (
	"context"

	"github.com/go-kit/kit/metrics"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(ClusterController) ClusterController

// LoggingMiddleware takes a logger as a dependency
// and returns a service Middleware.
func LoggingMiddleware(logger *zap.SugaredLogger) Middleware {
	return func(next ClusterController) ClusterController {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger *zap.SugaredLogger
	next   ClusterController
}

func (mw loggingMiddleware) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (res *proto.ClusterResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateCluster", "provider", req.Provider, "region", req.Region, "node", req.Node, "labels", req.Labels, "response", res, "error", err)
	}()
	return mw.next.CreateCluster(ctx, req)
}

func (mw loggingMiddleware) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (res *proto.ClusterDeleteResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "DeleteCluster", "name", req.ClusterName, "response", res, "error", err)
	}()
	return mw.next.DeleteCluster(ctx, req)
}

func (mw loggingMiddleware) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (res *proto.ClusterSpec, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "GetCluster", "name", req.ClusterName, "response", res, "error", err)
	}()
	return mw.next.GetCluster(ctx, req)
}

func (mw loggingMiddleware) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (res *proto.GetClustersResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "GetClusters", "provider", req.Provider, "region", "response", res, "error", err)
	}()
	return mw.next.GetClusters(ctx, req)
}

func (mw loggingMiddleware) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (res *proto.ClusterStatusResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "ClusterStatus", "name", req.ClusterName, "response", res, "error", err)
	}()
	return mw.next.ClusterStatus(ctx, req)
}

func (mw loggingMiddleware) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (res *proto.NodeSpawnResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "AddNode", "name", req.ClusterName, "nodespecinstance", req.NodeSpec.Instance, "response", res, "error", err)
	}()
	return mw.next.AddNode(ctx, req)
}

func (mw loggingMiddleware) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (res *proto.NodeDeleteResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "DeleteNode", "name", req.ClusterName, "nodegroupname", req.NodeGroupName, "response", res, "error", err)
	}()
	return mw.next.DeleteNode(ctx, req)
}

func (mw loggingMiddleware) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (res *proto.CreateVolumeResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateVolume", "volumetype", req.Volumetype, "size", req.Size, "region", req.Region, "provider", req.Provider, "snapshotid", req.Snapshotid, "response", res, "error", err)
	}()
	return mw.next.CreateVolume(ctx, req)
}

func (mw loggingMiddleware) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (res *proto.DeleteVolumeResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "DeleteVolume", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "response", res, "error", err)
	}()
	return mw.next.DeleteVolume(ctx, req)
}

func (mw loggingMiddleware) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (res *proto.CreateSnapshotResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateSnapshot", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "response", res, "error", err)
	}()
	return mw.next.CreateSnapshot(ctx, req)
}

func (mw loggingMiddleware) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (res *proto.CreateSnapshotAndDeleteResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateSnapshotAndDelete", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "response", res, "error", err)
	}()
	return mw.next.CreateSnapshotAndDelete(ctx, req)
}

func (mw loggingMiddleware) AddToken(ctx context.Context, req *proto.AddTokenRequest) (res *proto.AddTokenResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "AddToken", "response", res, "error", err)
	}()
	return mw.next.AddToken(ctx, req)
}

func (mw loggingMiddleware) GetToken(ctx context.Context, req *proto.GetTokenRequest) (res *proto.GetTokenResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "GetToken", "error", err)
	}()
	return mw.next.GetToken(ctx, req)
}

func (mw loggingMiddleware) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (res *proto.AddRoute53RecordResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "AddRoute53Record", "response", res, "error", err)
	}()
	return mw.next.AddRoute53Record(ctx, req)
}

func (mw loggingMiddleware) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (res *proto.RancherRegistrationResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "RegisterWithRancher", "response", res, "error", err)
	}()
	return mw.next.RegisterWithRancher(ctx, req)
}

// InstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
func InstrumentingMiddleware(ints metrics.Counter) Middleware {
	return func(next ClusterController) ClusterController {
		return instrumentingMiddleware{ints, next}
	}
}

type instrumentingMiddleware struct {
	ints metrics.Counter
	next ClusterController
}

func (mw instrumentingMiddleware) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	v, err := mw.next.CreateCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (res *proto.ClusterSpec, err error) {
	v, err := mw.next.GetCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (res *proto.GetClustersResponse, err error) {
	v, err := mw.next.GetClusters(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	v, err := mw.next.ClusterStatus(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	v, err := mw.next.DeleteCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	v, err := mw.next.AddNode(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	v, err := mw.next.DeleteNode(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	v, err := mw.next.CreateVolume(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	v, err := mw.next.DeleteVolume(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	v, err := mw.next.CreateSnapshot(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	v, err := mw.next.CreateSnapshotAndDelete(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	v, err := mw.next.AddToken(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	v, err := mw.next.GetToken(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	v, err := mw.next.AddRoute53Record(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) RegisterWithRancher(ctx context.Context, req *proto.RancherRegistrationRequest) (*proto.RancherRegistrationResponse, error) {
	v, err := mw.next.RegisterWithRancher(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}
