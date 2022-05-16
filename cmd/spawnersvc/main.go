package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/netbook-ai/interceptors"
	"github.com/oklog/oklog/pkg/group"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/gateway"
	"gitlab.com/netbook-devs/spawner-service/pkg/metrics"
	"gitlab.com/netbook-devs/spawner-service/pkg/service"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func startHttpServer(g *group.Group, config config.Config, logger *zap.SugaredLogger) {

	address := fmt.Sprintf("%s:%d", "", config.DebugPort)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorw("startHttpServer: failed to listen", "error", err)
		os.Exit(1)

	}

	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())

	g.Add(func() error {

		logger.Infow("startHttpServer", "transport", "debug/HTTP", "address", address)
		return http.Serve(listener, router)
	}, func(err error) {
		logger.Errorw("http-listener", "error", err)
		listener.Close()
	})
}

func startGRPCServer(g *group.Group, config config.Config, logger *zap.SugaredLogger) {

	address := fmt.Sprintf("%s:%d", "", config.Port)
	service := service.New(logger)
	grpcServer := gateway.New(service)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorw("startGRPCServer", "transport", "gRPC", "during", "Listen", "error", err)
		os.Exit(1)
	}

	interceptors := interceptors.NewInterceptor("spawnerservice",
		logger,
		interceptors.WithInterecptor(metrics.RPCInstrumentation()))

	g.Add(func() error {
		logger.Infow("startGRPCServer", "transport", "gRPC", "address", address)

		baseServer := grpc.NewServer(interceptors.Get())

		proto.RegisterSpawnerServiceServer(baseServer, grpcServer)
		return baseServer.Serve(listener)
	}, func(error) {
		logger.Errorw("startGRPCServer", "error", err)
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

	err := config.Load("./../../")

	if err != nil {
		log.Fatal("failed to load config", err)
	}

	var logger *zap.Logger

	//ENV value can be either prod or dev
	config := config.Get()
	if config.Env == "prod" || config.Env == "dev" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	var sugar = logger.Sugar()
	defer sugar.Sync()

	if err != nil {
		sugar.Errorw("error loading config", "error", err.Error())
	}
	var g group.Group

	startHttpServer(&g, config, sugar)
	startGRPCServer(&g, config, sugar)
	startSignalHandler(&g)

	sugar.Infow("main", "exit", g.Run())
}
