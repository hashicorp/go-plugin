This is a command tool to generate proto3 to grpc plugin golang source code.

# How to use
## 1.download gogo proto:
```shell
mkdir -p ${HOME}/code/github.com/gogo/
cd ${HOME}/code/github.com/gogo/
git clone https://github.com/gogo/protobuf.git
```

## 2.add a proto3 idl file:
see example: [./proto/my_test_grpc_plugin.proto]
* define Request and Response message format
* define a service
* use protoc to generate XX.pb.go files
```shell
make pb proto_path="${HOME}/code/"
```
