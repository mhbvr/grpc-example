module github.com/mhbvr/grpc-example

go 1.20

require (
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.41.1
	go.opentelemetry.io/contrib/zpages v0.41.1
	go.opentelemetry.io/otel v1.15.1
	go.opentelemetry.io/otel/sdk v1.15.1
	go.opentelemetry.io/otel/trace v1.15.1
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/go-chi/chi/v5 v5.0.7 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.opentelemetry.io/otel/metric v0.38.1 // indirect
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/rantav/go-grpc-channelz v0.0.3
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
)

replace go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => github.com/mhbvr/opentelemetry-go-contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.0.3
