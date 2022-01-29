package spawnerservice

import (
	"context"

	pb "gitlab.com/netbook-devs/spawner-service/pb"

	"github.com/go-kit/kit/metrics"
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

func (mw loggingMiddleware) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (res *pb.ClusterResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateCluster", "provider", req.Provider, "region", req.Region, "node", req.Node, "labels", req.Labels, "response", res, "error", err)
	}()
	return mw.next.CreateCluster(ctx, req)
}

func (mw loggingMiddleware) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (res *pb.ClusterDeleteResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "DeleteCluster", "name", req.ClusterName, "response", res, "error", err)
	}()
	return mw.next.DeleteCluster(ctx, req)
}

func (mw loggingMiddleware) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (res *pb.ClusterSpec, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "GetCluster", "name", req.ClusterName, "response", res, "error", err)
	}()
	return mw.next.GetCluster(ctx, req)
}

func (mw loggingMiddleware) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (res *pb.GetClustersResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "GetClusters", "provider", req.Provider, "region", req.Region, "scope", req.Scope, "response", res, "error", err)
	}()
	return mw.next.GetClusters(ctx, req)
}

func (mw loggingMiddleware) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (res *pb.ClusterStatusResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "ClusterStatus", "name", req.ClusterName, "response", res, "error", err)
	}()
	return mw.next.ClusterStatus(ctx, req)
}

func (mw loggingMiddleware) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (res *pb.NodeSpawnResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "AddNode", "name", req.ClusterName, "nodespecinstance", req.NodeSpec.Instance, "response", res, "error", err)
	}()
	return mw.next.AddNode(ctx, req)
}

func (mw loggingMiddleware) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (res *pb.NodeDeleteResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "DeleteNode", "name", req.ClusterName, "nodegroupname", req.NodeGroupName, "response", res, "error", err)
	}()
	return mw.next.DeleteNode(ctx, req)
}

func (mw loggingMiddleware) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (res *pb.CreateVolumeResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateVolume", "volumetype", req.Volumetype, "size", req.Size, "region", req.Region, "provider", req.Provider, "snapshotid", req.Snapshotid, "response", res, "error", err)
	}()
	return mw.next.CreateVolume(ctx, req)
}

func (mw loggingMiddleware) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (res *pb.DeleteVolumeResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "DeleteVolume", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "response", res, "error", err)
	}()
	return mw.next.DeleteVolume(ctx, req)
}

func (mw loggingMiddleware) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (res *pb.CreateSnapshotResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateSnapshot", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "response", res, "error", err)
	}()
	return mw.next.CreateSnapshot(ctx, req)
}

func (mw loggingMiddleware) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (res *pb.CreateSnapshotAndDeleteResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "CreateSnapshotAndDelete", "volumeid", req.Volumeid, "region", req.Region, "provider", req.Provider, "response", res, "error", err)
	}()
	return mw.next.CreateSnapshotAndDelete(ctx, req)
}

func (mw loggingMiddleware) AddToken(ctx context.Context, req *pb.AddTokenRequest) (res *pb.AddTokenResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "AddToken", "response", res, "error", err)
	}()
	return mw.next.AddToken(ctx, req)
}

func (mw loggingMiddleware) GetToken(ctx context.Context, req *pb.GetTokenRequest) (res *pb.GetTokenResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "GetToken", "error", err)
	}()
	return mw.next.GetToken(ctx, req)
}

func (mw loggingMiddleware) AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (res *pb.AddRoute53RecordResponse, err error) {
	defer func() {
		mw.logger.Infow("spawnerservice", "method", "AddRoute53Record", "response", res, "error", err)
	}()
	return mw.next.AddRoute53Record(ctx, req)
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

func (mw instrumentingMiddleware) CreateCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	v, err := mw.next.CreateCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) GetCluster(ctx context.Context, req *pb.GetClusterRequest) (res *pb.ClusterSpec, err error) {
	v, err := mw.next.GetCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) GetClusters(ctx context.Context, req *pb.GetClustersRequest) (res *pb.GetClustersResponse, err error) {
	v, err := mw.next.GetClusters(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) ClusterStatus(ctx context.Context, req *pb.ClusterStatusRequest) (*pb.ClusterStatusResponse, error) {
	v, err := mw.next.ClusterStatus(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteCluster(ctx context.Context, req *pb.ClusterDeleteRequest) (*pb.ClusterDeleteResponse, error) {
	v, err := mw.next.DeleteCluster(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddNode(ctx context.Context, req *pb.NodeSpawnRequest) (*pb.NodeSpawnResponse, error) {
	v, err := mw.next.AddNode(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteNode(ctx context.Context, req *pb.NodeDeleteRequest) (*pb.NodeDeleteResponse, error) {
	v, err := mw.next.DeleteNode(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	v, err := mw.next.CreateVolume(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) DeleteVolume(ctx context.Context, req *pb.DeleteVolumeRequest) (*pb.DeleteVolumeResponse, error) {
	v, err := mw.next.DeleteVolume(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	v, err := mw.next.CreateSnapshot(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) CreateSnapshotAndDelete(ctx context.Context, req *pb.CreateSnapshotAndDeleteRequest) (*pb.CreateSnapshotAndDeleteResponse, error) {
	v, err := mw.next.CreateSnapshotAndDelete(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddToken(ctx context.Context, req *pb.AddTokenRequest) (*pb.AddTokenResponse, error) {
	v, err := mw.next.AddToken(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	v, err := mw.next.GetToken(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}

func (mw instrumentingMiddleware) AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (*pb.AddRoute53RecordResponse, error) {
	v, err := mw.next.AddRoute53Record(ctx, req)
	mw.ints.Add(float64(1))
	return v, err
}
