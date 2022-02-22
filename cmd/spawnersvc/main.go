package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/gogo/status"
	"github.com/oklog/oklog/pkg/group"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwntransport"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func getMeth(info *grpc.UnaryServerInfo) string {
	splits := strings.Split(info.FullMethod, "/")
	return splits[len(splits)-1]

}

func getUnaryRecoveryInterceptors() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				//log error details and stack trace
				fmt.Printf("%s panicked with %s", info.FullMethod, r.(error))
				fmt.Printf("StackTrace: %s", string(debug.Stack()))
				methodName := getMeth(info)
				err = status.Errorf(codes.Internal, "%v in call to method '%s'", r, methodName)
			}
		}()

		resp, err := handler(ctx, req)
		panicked = false
		return resp, err
	}
}

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
	// {
	// 	// The HTTP listener mounts the Go kit HTTP handler we created.
	// 	httpListener, err := net.Listen("tcp", *httpAddr)
	// 	if err != nil {
	// 		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
	// 		os.Exit(1)
	// 	}
	// 	g.Add(func() error {
	// 		logger.Log("transport", "HTTP", "addr", *httpAddr)
	// 		return http.Serve(httpListener, httpHandler)
	// 	}, func(error) {
	// 		httpListener.Close()
	// 	})
	// }
	// The gRPC listener mounts the Go kit gRPC server we created.
	grpcAddr := fmt.Sprintf("%s:%d", "", config.Port)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		sugar.Errorw("error in grpcListener", "transport", "gRPC", "during", "Listen", "err", err)
		os.Exit(1)
	}
	g.Add(func() error {
		sugar.Infow("in main", "transport", "gRPC", "grpcAddr", grpcAddr)
		// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
		// the here demonstrated zipkin tracing middleware.
		baseServer := grpc.NewServer(grpc.ChainUnaryInterceptor(kitgrpc.Interceptor, getUnaryRecoveryInterceptors()))
		proto.RegisterSpawnerServiceServer(baseServer, grpcServer)
		return baseServer.Serve(grpcListener)
	}, func(error) {
		grpcListener.Close()
	})
	// This function just sits and waits for ctrl-C.
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
