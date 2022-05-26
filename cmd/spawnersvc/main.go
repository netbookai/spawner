package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/netbook-ai/interceptors"
	"github.com/netbookai/log"
	"github.com/netbookai/log/loggers/zap"
	"github.com/oklog/oklog/pkg/group"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/gateway"
	"gitlab.com/netbook-devs/spawner-service/pkg/metrics"
	"gitlab.com/netbook-devs/spawner-service/pkg/service"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"google.golang.org/grpc"
)

func startHttpServer(ctx context.Context, g *group.Group, config config.Config, logger log.Logger) {

	address := fmt.Sprintf("%s:%d", "", config.DebugPort)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error(ctx, "startHttpServer: failed to listen", "error", err)
		os.Exit(1)

	}

	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())

	g.Add(func() error {
		logger.Info(ctx, "startHttpServer", "transport", "debug/HTTP", "address", address)
		return http.Serve(listener, router)
	}, func(err error) {
		logger.Error(ctx, "http-listener", "error", err)
		listener.Close()
	})
}

func startGRPCServer(ctx context.Context, g *group.Group, config config.Config, logger log.Logger) {

	address := fmt.Sprintf("%s:%d", "", config.Port)
	service := service.New(logger)
	grpcServer := gateway.New(service)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error(ctx, "startGRPCServer", "transport", "gRPC", "during", "Listen", "error", err)
		os.Exit(1)
	}

	interceptors := interceptors.NewInterceptor("spawnerservice",
		logger,
		interceptors.WithInterecptor(metrics.RPCInstrumentation()))

	g.Add(func() error {
		logger.Info(ctx, "startGRPCServer", "transport", "gRPC", "address", address)

		baseServer := grpc.NewServer(interceptors.Get())

		proto.RegisterSpawnerServiceServer(baseServer, grpcServer)
		return baseServer.Serve(listener)
	}, func(error) {
		logger.Error(ctx, "startGRPCServer", "error", err)
		listener.Close()
	})

}

func startSignalHandler(g *group.Group) {

	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})
}

func main() {

	ctx := context.Background()
	err := config.Load(".")
	logger := log.NewLogger(zap.NewLogger())

	if err != nil {
		logger.Error(ctx, "failed to load config", "error", err)
		return
	}

	//ENV value can be either prod or dev
	config := config.Get()

	if err != nil {
		logger.Error(ctx, "error loading config", "error", err.Error())
	}
	var g group.Group

	startHttpServer(ctx, &g, config, logger)
	startGRPCServer(ctx, &g, config, logger)
	startSignalHandler(&g)

	logger.Info(ctx, "main", "exit", g.Run())
}
