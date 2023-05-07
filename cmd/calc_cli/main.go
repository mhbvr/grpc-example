package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mhbvr/grpc-example/pkg/eval"
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
	if len(os.Args) < 2 {
		log.Fatalf("Need to provide an expression for evaluation")
	}

	expr, err := eval.Parse(os.Args[1])
	if err != nil {
		log.Fatalf("incorrect expression, %v", err)
	}

	var env eval.Env = make(map[eval.Var]float64)

	for _, s := range os.Args[2:] {
		variable, value, err := parseValue(s)
		if err != nil {
			log.Fatalf("incorrect variable, %v", err)
		}
		env[eval.Var(variable)] = value
	}

	result := expr.Eval(env)
	fmt.Printf("Result: %v\n", result)
}
