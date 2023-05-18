package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	pb "github.com/mhbvr/grpc-example/pkg/calc_protos"
	"github.com/mhbvr/grpc-example/pkg/eval"
	"google.golang.org/grpc"
	channelzservice "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	//"google.golang.org/protobuf/types/known/"

	channelz "github.com/rantav/go-grpc-channelz"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/zpages"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	port     = flag.Int("port", 5050, "The server gRPC port")
	httpPort = flag.Int("http_port", 8080, "The server HTTP diagnostic port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedCalcServer
}

// Implementation of Compute RC
func (s *server) Compute(ctx context.Context, in *pb.ComputeRequest) (*pb.ComputeReply, error) {
	span := trace.SpanFromContext(ctx)
	log.Printf("ComputeRequest: %v", in)

	span.AddEvent("start parsing")
	expr, err := eval.Parse(in.Expression)
	if err != nil {
		log.Printf("incorrect expression, %v", err)
		return nil, err
	}

	span.AddEvent("start evaliation")
	var env eval.Env = make(map[eval.Var]float64)
	for _, v := range in.Vars {
		env[eval.Var(v.Name)] = v.Value
	}

	return &pb.ComputeReply{Result: expr.Eval(env)}, nil
}

func (s *server) StreamCompute(stream pb.Calc_StreamComputeServer) error {
	span := trace.SpanFromContext(stream.Context())
	var expr eval.Expr
	for {
		in, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Printf("gRPC stream recv error, %v", err)
			return err
		}

		// First message should have non empty expression to compute
		// It is possible to set another expression in the next messages
		if in.Expression == "" && expr == nil {
			return status.Error(codes.InvalidArgument, "initial messase without expression")
		}

		if in.Expression != "" {
			span.AddEvent("start parsing expression")
			var err error
			expr, err = eval.Parse(in.Expression)

			if err != nil {
				return status.Errorf(codes.InvalidArgument, "can not parse expression %v", err)
			}
		}

		var env eval.Env = make(map[eval.Var]float64)
		for _, v := range in.Vars {
			env[eval.Var(v.Name)] = v.Value
		}

		stream.Send(&pb.ComputeReply{Result: expr.Eval(env)})
	}
}

func main() {
	flag.Parse()

	// Create default TraceProvider with zpagez trace processor
	spanProc := zpages.NewSpanProcessor()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(spanProc))

	otel.SetTracerProvider(tp)

	grpcBindAddress := fmt.Sprintf(":%d", *port)

	lis, err := net.Listen("tcp", grpcBindAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()), grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))

	// Service for expression evaluation
	pb.RegisterCalcServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	// Register channelz service
	channelzservice.RegisterChannelzServiceToServer(s)

	http.Handle("/tracez", zpages.NewTracezHandler(spanProc))
	http.Handle("/", channelz.CreateHandler("/", grpcBindAddress))

	// Listen and serve HTTP for the default serve mux
	httpBindAddress := fmt.Sprintf(":%d", *httpPort)

	httpListener, err := net.Listen("tcp", httpBindAddress)
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(httpListener, nil)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
