package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	pb "github.com/mhbvr/grpc-example/pkg/calc_protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:5050", "the address of Calc server")
)

func parseValue(s string) (string, float64, error) {
	fields := strings.Split(s, "=")
	if len(fields) != 2 {
		return "", 0.0, fmt.Errorf("can not parse variable, %v", s)
	}
	value, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return "", 0.0, fmt.Errorf("can not parse float value, %v", fields[1])
	}

	if len(fields[0]) == 0 {
		return "", 0.0, fmt.Errorf("empty variable name")
	}
	return fields[0], value, nil
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("Need to provide an expression for evaluation")
	}

	request := &pb.ComputeRequest{Expression: flag.Args()[0]}

	vars := make([]*pb.Variable, 0)

	for _, s := range flag.Args()[1:] {
		variable, value, err := parseValue(s)
		if err != nil {
			log.Fatalf("incorrect variable, %v", err)
		}
		vars = append(vars, &pb.Variable{Name: variable, Value: value})
	}

	request.Vars = vars

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCalcClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Compute(ctx, request)
	if err != nil {
		log.Fatalf("could not compute: %v", err)
	}

	fmt.Printf("Result: %v\n", r.GetResult())
}
