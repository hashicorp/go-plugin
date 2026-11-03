
proto_path="${HOME}/code/"

.PHONY: pb plugin gen
pb:
	protoc -I=. -I=${proto_path} --gogofaster_out=plugins=grpc:. proto/*.proto

code_gen: cmd/code_gen.go cmd/go.mod
	cd cmd && make build && cd .. && mv cmd/code_gen ./
	#go mod tidy && go build -o code_gen cmd/code_gen.go

plugin:
	./code_gen -proto_path=${proto_path} -source_proto=./proto/my_test_grpc_plugin.proto -target_path=./

gen: pb code_gen plugin
	@echo ok
