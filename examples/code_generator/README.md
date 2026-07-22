This is a command tool to generate proto3 to grpc plugin golang source code.

# How to use
## 0.install protoc
```shell
brew install protobuf
```

## 1.download gogo proto:
```shell
mkdir -p ${HOME}/code/github.com/gogo/
cd ${HOME}/code/github.com/gogo/
git clone https://github.com/gogo/protobuf.git
```
And install protoc plugin of gogo proto:
```shell
go install github.com/gogo/protobuf/protoc-gen-gogo
```

## 2.add a proto3 idl file:
see example: [./proto/my_test_grpc_plugin.proto]
* define Request and Response message format
* define a service
  - use extension to add a plugin name:
  - `option (plugin_name) = "my_plugin_1";`
  - It will use service name when plugin name not set
* use protoc to generate XX.pb.go files
```shell
make pb proto_path="${HOME}/code/"
```

## 3.generate code
```shell
make gen proto_path="${HOME}/code/"
```

see current directory `github.com/...`.

# Test plugin
```shell
cd examples/code_generator/github.com/hashicorp/go-plugin/examples/my_test_grpc_plugin_callee
go mod tidy -compat=1.17 && go build
# add code to callee_test.go
go test
```
