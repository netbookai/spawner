module gitlab.com/netbook-devs/spawner-service

go 1.17

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/client-go => github.com/rancher/client-go v0.21.0-rancher.1
)

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rancher/wrangler v0.8.6-0.20210819203859-0babd42fbad8 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/streadway/handy v0.0.0-20200128134331-0f66f006fb2e // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.21.3 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
)

require (
	github.com/go-kit/kit v0.11.0
	github.com/oklog/oklog v0.3.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/rancher/norman v0.0.0-20210709145327-afd06f533ca3
	github.com/rancher/rancher/pkg/client v0.0.0-20210928015834-4df101e320c3
	github.com/sony/gobreaker v0.4.1
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)
