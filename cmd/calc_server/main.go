package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/mhbvr/grpc-example/pkg/calc_protos"
	"github.com/mhbvr/grpc-example/pkg/eval"
	"google.golang.org/grpc"
	channelzservice "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"

	channelz "github.com/rantav/go-grpc-channelz"
)

var (
	port = flag.Int("port", 5050, "The server gRPC port")
	httpPort = flag.Int("http_port", 8080, "The server HTTP diagnostic port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedCalcServer
}

// Implementation of Compute RC
func (s *server) Compute(ctx context.Context, in *pb.ComputeRequest) (*pb.ComputeReply, error) {
	log.Printf("ComputeRequest: %v", in)

	expr, err := eval.Parse(in.Expression)
	if err != nil {
		log.Printf("incorrect expression, %v", err)
		return nil, err
	}

	var env eval.Env = make(map[eval.Var]float64)
	for _, v := range in.Vars {
		env[eval.Var(v.Name)] = v.Value
	}

	return &pb.ComputeReply{Result: expr.Eval(env)}, nil
}

func main() {
	flag.Parse()

	grpcBindAddress := fmt.Sprintf(":%d", *port)

	lis, err := net.Listen("tcp", grpcBindAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Service for expression evaluation 
	pb.RegisterCalcServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	// Register channelz service
	channelzservice.RegisterChannelzServiceToServer(s) 
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
