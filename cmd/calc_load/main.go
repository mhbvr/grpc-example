package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/mhbvr/grpc-example/pkg/calc_protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr      = flag.String("addr", "localhost:5050", "the address of Calc server")
	duration  = flag.Duration("duration", 10*time.Second, "duration of load test")
	timeout   = flag.Duration("timeout", time.Second, "timeout for one request")
	rate      = flag.Float64("rate", 1, "requests per second")
	dry_run   = flag.Bool("dry_run", false, "do not send actual requests, just start goroutine to get max possible request rate")
	streaming = flag.Bool("streaming", false, "use one bidirectional gRPC stream")
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

func runLoad(c pb.CalcClient, expression string, vars []*pb.Variable, okChan, errChan chan struct{}) {
	request := &pb.ComputeRequest{Expression: expression, Vars: vars}
	var wg sync.WaitGroup

	var count int

	// For each request start separate goroutine
	for start := time.Now(); time.Since(start) < *duration; {

		if float64(count)/time.Since(start).Seconds() > *rate {
			continue
		}

		count++
		wg.Add(1)
		// Start sender goroutine
		go func() {
			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			defer cancel()

			var err error
			if !*dry_run {
				_, err = c.Compute(ctx, request)
			}

			if err == nil {
				okChan <- struct{}{}
			} else {
				errChan <- struct{}{}
			}

			wg.Done()
		}()
	}

	// Wait for completion of all sender gorouties
	wg.Wait()
}

func runStreamingLoad(c pb.CalcClient, expression string, vars []*pb.Variable, okChan, errChan chan struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	stream, err := c.StreamCompute(ctx)
	if err != nil {
		log.Fatalf("can not establich streaming connect, %v", err)
	}

	// Initial message
	err = stream.Send(&pb.ComputeRequest{Expression: expression})
	// TODO: handle io.EOF
	if err != nil {
		log.Fatalf("can not send first streaming message, %v", err)
	}

	var wg sync.WaitGroup

	// Sender
	wg.Add(1)
	go func() {
		var err error
		var count int
		for start := time.Now(); time.Since(start) < *duration; {

			if float64(count)/time.Since(start).Seconds() > *rate {
				continue
			}

			count++

			if !*dry_run {
				err = stream.Send(&pb.ComputeRequest{Vars: vars})
			}

			if err == io.EOF {
				log.Println("stream aborted")
				wg.Done()
				return
			}

			if err != nil {
				log.Fatalf("can not send streaming message, %v", err)
			}

		}

		err = stream.CloseSend()
		if err != nil {
			log.Fatalf("can not close stream, %v", err)
		}
		wg.Done()
	}()

	// Receiver
	if !*dry_run {
		wg.Add(1)
		go func() {
			var err error
			for {
				_, err = stream.Recv()

				if err == io.EOF {
					break
				}

				if err != nil {
					log.Printf("gRPC stream error, %v", err)
					break
				}

				okChan <- struct{}{}
			}
			wg.Done()
		}()
	}

	// Wait for complition for sender and receiver
	wg.Wait()
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("Need to provide an expression for evaluation")
	}

	vars := make([]*pb.Variable, 0)
	for _, s := range flag.Args()[1:] {
		variable, value, err := parseValue(s)
		if err != nil {
			log.Fatalf("incorrect variable, %v", err)
		}
		vars = append(vars, &pb.Variable{Name: variable, Value: value})
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCalcClient(conn)

	// Start goroutine to collect results
	var success, errors int
	okChan := make(chan struct{})
	errChan := make(chan struct{})
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-okChan:
				success++
			case <-errChan:
				errors++
			case <-done:
				break
			}
		}
	}()

	if *streaming {
		runStreamingLoad(c, flag.Args()[0], vars, okChan, errChan)
	} else {
		runLoad(c, flag.Args()[0], vars, okChan, errChan)
	}

	done <- struct{}{}

	fmt.Printf("Ok: %v Errors: %v, Rate: %v req/sec\n", success, errors, float64(success+errors)/duration.Seconds())
}
