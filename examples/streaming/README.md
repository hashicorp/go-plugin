## To re-generate protobuf definitions

Install protobuf tooling

```
brew install protobuf
```
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

generate files

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative streamer.proto
```

## To execute

Build the plugin

```
go build -o ./plugin/streamer ./plugin
```

generate a file to stream

```
head -c 1000000000 </dev/urandom > myfile
```

launch

```
go run main.go myfile
```