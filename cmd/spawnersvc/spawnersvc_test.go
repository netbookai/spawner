package main

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwntransport"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

var logger, _ = zap.NewDevelopment()
var sugar = logger.Sugar()

func init() {

	config, err := config.Load("../../")
	if err != nil {
		sugar.Errorw("error loading config", "error", err.Error())
	}

	// Create the (sparse) metrics we'll use in the service. They, too, are
	// dependencies that we pass to components that use them.
	var ints metrics.Counter
	{
		// Business-level metrics.
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "spawnerservice",
			Subsystem: "spawnerservice",
			Name:      "integers_summed",
			Help:      "Total number of method calls",
		}, []string{})
	}

	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "spawnerservice",
			Subsystem: "spawnerservice",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		service   = spawnerservice.New(sugar, &config, ints)
		endpoints = spwnendpoint.New(service, sugar, duration)
		// httpHandler    = spwntransport.NewHTTPHandler(endpoints, tracer, zipkinTracer, logger)
		grpcServer = spwntransport.NewGRPCServer(endpoints, sugar)
	)

	lis = bufconn.Listen(bufSize)
	baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterSpawnerServiceServer(baseServer, grpcServer)
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			sugar.Errorw("Server exited with error", "error", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestClusterStatus(t *testing.T) {

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		sugar.Errorw("error connecting to server", "error", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	svc := spwntransport.NewGRPCClient(conn, zap.NewNop().Sugar())

	clusterStatusReq := &pb.ClusterStatusRequest{
		ClusterName: "infra-test",
	}

	resp, _ := svc.ClusterStatus(context.Background(), clusterStatusReq)

	if resp.Error != "" {
		t.Errorf("error in calling cluster status api: %s", resp.Error)
	}
	if resp.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", resp.Status)
	}
}
