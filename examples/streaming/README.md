# gRPC streaming Example

This example builds a plugin & client which can stream larger amount of data
between them while staying below reasonable message size limits of the gRPC
protocol.

## To execute

Build the plugin

```
go build -o ./plugin/streamer ./plugin
```

launch client

```
go run main.go myfile
```

## To re-generate protobuf definitions

Install protobuf tooling

```
brew install protobuf
```
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
```

generate files

```
cd proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative streamer.proto
```