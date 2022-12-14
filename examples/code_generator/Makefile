
proto_path="${HOME}/code/"

.PHONY: pb
pb:
	protoc -I=. -I=${proto_path} --gogofaster_out=plugins=grpc:. proto/*.proto

