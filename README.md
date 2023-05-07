# gRPC sandbox
Toy server and client apps to play with gRPC.

## Examples

- `calc_cli` - program to calculate results of mathimatical expression localy
- `calc_server` - gRPC server for expression calculation
- `calc_client` - gRPC client for the server
- `calc_load` - gRPC client sends calculation requests with constant rate

## Features

- [gRPC reflection](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md) for [grpcurl](https://github.com/fullstorydev/grpcurl)

## TODO

- gRPC [channelz](https://github.com/grpc/proposal/blob/master/A14-channelz.md) introspection. Several [GUI](https://github.com/grpc/grpc-experiments/tree/master/gdebug) [options](https://github.com/rantav/go-grpc-channelz) avaiable
- Prometheus metrics
- tracez page
- profiling and tracing gRPC calls

## License

These apps use eval package from "The Go Programming Language" book; see http://www.gopl.io. Source code examples from the book are available from [gopl.io repo](https://github.com/adonovan/gopl.io) and licensed under a [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](http://creativecommons.org/licenses/by-nc-sa/4.0/) This repo is using the same license.

![Creative Commons License](https://i.creativecommons.org/l/by-nc-sa/4.0/88x31.png)
