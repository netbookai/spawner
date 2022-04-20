package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var requestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "grpc_request_total",
		Help: "Number of gRPC calls",
	},
	[]string{"method"},
)

func init() {
	prometheus.Register(requestCounter)
}

//IncRequest incerement total request counter
func IncRequest(method string) {
	requestCounter.WithLabelValues(method).Inc()
}
