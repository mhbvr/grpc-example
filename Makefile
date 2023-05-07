all: build 

proto: 
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/calc_protos/calc.proto

build: clean proto
	go build -o bin/calc_cli cmd/calc_cli/main.go
	go build -o bin/calc_server cmd/calc_server/main.go
	go build -o bin/calc_client cmd/calc_client/main.go	

clean:
	rm -rf bin
	rm -f pkg/calc_protos/*.pb.go