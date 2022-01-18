package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/oklog/oklog/pkg/group"
	"github.com/pkg/errors"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"gitlab.com/netbook-devs/spawner-service/pb"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwnendpoint"
	"gitlab.com/netbook-devs/spawner-service/pkg/spwntransport"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {

	conf, err := config.Load("../../")
	if err != nil {
		fmt.Errorf("error loading config", "error", err.Error())
	}

	var logger *zap.Logger
	if conf.Env == "development" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	sugar := logger.Sugar()
	defer sugar.Sync()

	//TODO: move to monitoring package
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
		service   = spawnerservice.New(sugar, &conf, ints)
		endpoints = spwnendpoint.New(service, sugar, duration)
		// httpHandler    = spwntransport.NewHTTPHandler(endpoints, tracer, zipkinTracer, logger)
		grpcServer = spwntransport.NewGRPCServer(endpoints, sugar)
	)

	// Now we're to the part of the func main where we want to start actually
	// running things, like servers bound to listeners to receive connections.
	//
	// The method is the same for each component: add a new actor to the group
	// struct, which is a combination of 2 anonymous functions: the first
	// function actually runs the component, and the second function should
	// interrupt the first function and cause it to return. It's in these
	// functions that we actually bind the Go kit server/handler structs to the
	// concrete transports and run them.
	//
	// Putting each component into its own block is mostly for aesthetics: it
	// clearly demarcates the scope in which each listener/socket may be used.
	var g group.Group
	// The debug listener mounts the http.DefaultServeMux, and serves up
	// stuff like the Prometheus metrics route, the Go debug and profiling
	// routes, and so on.
	debugAddress := fmt.Sprintf("%s:%d", "", conf.DebugPort)
	debugListener, err := net.Listen("tcp", debugAddress)
	if err != nil {
		sugar.Errorw("error in debugListener", "transport", "debug/HTTP", "during", "Listen", "error", err)
		os.Exit(1)
	}
	g.Add(func() error {
		err = http.Serve(debugListener, http.DefaultServeMux)
		return errors.Wrap(err, "error in debugListener")

	}, func(error) {
		debugListener.Close()
	})

	// The gRPC listener mounts the Go kit gRPC server we created.

	grpcAddress := fmt.Sprintf("%s:%d", "", conf.Port)
	grpcListener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		sugar.Errorw("error in grpcListener", "transport", "gRPC", "during", "Listen", "err", err)
		os.Exit(1)
	}
	g.Add(func() error {
		sugar.Infow("in main", "transport", "gRPC", "grpcAddr", grpcAddress)
		// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
		// the here demonstrated zipkin tracing middleware.
		baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
		pb.RegisterSpawnerServiceServer(baseServer, grpcServer)
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
