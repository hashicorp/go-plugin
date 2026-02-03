# gRPC streaming Example

This example builds a plugin & client which can stream a large amount of data
between them while staying below reasonable message size limits of the gRPC
protocol.

> Note: [hashicorp/go-plugin sets an upper limit on message size](https://github.com/hashicorp/go-plugin/blob/d0d30899ca2d91b0869cb73db95afca180e769cf/grpc_client.go#L39-L41). At time of writing, that value is `math.MaxInt32` bytes, or approximately 2GB.

## To execute

Build the plugin

```
go build -o ./plugin/streamer ./plugin
```

Finally launch the client:

```
go run main.go myfile
```

The client will first write data to the streamer plugin, and then the client will read that
data back from the plugin. The plugin writes the data it receives in a file called `myfile`,
due to the argument passed to the client above.

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
