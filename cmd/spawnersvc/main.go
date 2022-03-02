package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/netbook-ai/interceptors"
	"github.com/oklog/oklog/pkg/group"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwntransport"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {

	config, err := config.Load(".")

	if err != nil {
		log.Fatal("failed to load config", err)
	}

	var logger *zap.Logger

	//ENV value can be either prod or dev
	if config.Env == "prod" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	var sugar = logger.Sugar()
	defer sugar.Sync()

	if err != nil {
		sugar.Errorw("error loading config", "error", err.Error())
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

	service := spawnerservice.New(sugar, &config)
	endpoints := spwnendpoint.New(service, sugar, duration)
	grpcServer := spwntransport.NewGRPCServer(endpoints, sugar)

	var g group.Group
	debugAddr := fmt.Sprintf("%s:%d", "", config.DebugPort)
	debugListener, err := net.Listen("tcp", debugAddr)
	if err != nil {
		sugar.Errorw("error in debugListener", "transport", "debug/HTTP", "during", "Listen", "error", err)
		os.Exit(1)
	}
	g.Add(func() error {
		sugar.Infow("error in debugListener", "transport", "debug/HTTP", "debugAddr", debugAddr)
		return http.Serve(debugListener, http.DefaultServeMux)
	}, func(error) {
		debugListener.Close()
	})

	grpcAddr := fmt.Sprintf("%s:%d", "", config.Port)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		sugar.Errorw("error in grpcListener", "transport", "gRPC", "during", "Listen", "err", err)
		os.Exit(1)
	}
	g.Add(func() error {
		sugar.Infow("in main", "transport", "gRPC", "grpcAddr", grpcAddr)
		baseServer := grpc.NewServer(interceptors.GetInterceptors("spawnerservice", sugar))
		proto.RegisterSpawnerServiceServer(baseServer, grpcServer)
		return baseServer.Serve(grpcListener)
	}, func(error) {
		grpcListener.Close()
	})

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
	sugar.Infow("main", "exit", g.Run())
}
