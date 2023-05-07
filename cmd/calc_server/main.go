package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/mhbvr/grpc-example/pkg/calc_protos"
	"github.com/mhbvr/grpc-example/pkg/eval"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 5050, "The server gRPC port")
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCalcServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
