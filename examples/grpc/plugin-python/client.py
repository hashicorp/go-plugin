
import kv_pb2
import kv_pb2_grpc


import grpc



def run():
    # NOTE(gRPC Python Team): .close() is possible on a channel and should be
    # used in circumstances in which the with statement does not fit the needs
    # of the code.
    with grpc.insecure_channel('localhost:1234') as channel:
        stub = kv_pb2_grpc.KVStub(channel)
        stub.Put(kv_pb2.PutRequest(key="cc",value=b"1"))
        response = stub.Get(kv_pb2.GetRequest(key="cc"))

    print("Greeter client received: " + str(response.value))

if __name__ == '__main__':
    run()
