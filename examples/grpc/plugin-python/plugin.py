from concurrent import futures
import sys
import time

import grpc

import kv_pb2
import kv_pb2_grpc

class KVServicer(kv_pb2_grpc.KVServicer):
    """Implementation of KV service."""

    def Get(self, request, context):
        filename = "kv_"+request.key
        with open(filename, 'r') as f:
            result = kv_pb2.GetResponse()
            result.value = f.read()
            return result

    def Put(self, request, context):
        filename = "kv_"+request.key
        value = "{0}\n\nWritten from plugin-python".format(request.value)
        with open(filename, 'w') as f:
            f.write(value)

        return kv_pb2.Empty()

def serve():
    # Start the server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    kv_pb2_grpc.add_KVServicer_to_server(KVServicer(), server)
    server.add_insecure_port(':1234')
    server.start()

    # Output information
    print("1|1|tcp|127.0.0.1:1234|grpc")
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
