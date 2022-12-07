import grpc
import logging
import sys
import time
from concurrent import futures
from grpc_health.v1 import health_pb2, health_pb2_grpc
from grpc_health.v1.health import HealthServicer
from io import StringIO
from queue import Queue
import queue
import grpc_stdio_pb2
import grpc_stdio_pb2_grpc
import kv_pb2
import kv_pb2_grpc
from logging.handlers import QueueHandler,QueueListener


class Logger:
    def __init__(self):
        self.stream = StringIO()  #
        que = Queue(-1)  # no limit on size
        self.queue_handler = QueueHandler(que)
        self.handler = logging.StreamHandler()
        self.listener = QueueListener(que, self.handler)
        self.log = logging.getLogger('python-plugin')
        self.log.setLevel(logging.DEBUG)
        self.logFormatter = logging.Formatter('%(asctime)s %(levelname)s  %(name)s %(pathname)s:%(lineno)d - %('
                                         'message)s')
        self.handler.setFormatter(self.logFormatter)
        for handler in self.log.handlers:
            self.log.removeHandler(handler)
        self.log.addHandler(self.queue_handler)
        self.listener.start()

    def __del__(self):
        self.listener.stop()

    def read(self):
        self.handler.flush()
        ret = self.logFormatter.format(self.listener.queue.get()) + "\n"
        return ret.encode("utf-8")


logger = Logger()
log = logger.log


class KVServicer(kv_pb2_grpc.KVServicer):
    """Implementation of KV service."""

    def Get(self, request, context):
        filename = "kv_" + request.key
        import time
        with open(filename, 'r+b') as f:
            result = kv_pb2.GetResponse()
            result.value = f.read()
            log.info(133)
            time.sleep(10)
            log.info(3333)
            time.sleep(10)


        return result

    def Put(self, request, context):
        filename = "kv_" + request.key
        value = "{0}\n\nWritten from plugin-python".format(request.value)
        with open(filename, 'w') as f:
            f.write(value)

        return kv_pb2.Empty()


class StdioService(grpc_stdio_pb2_grpc.GRPCStdioServicer):
    def __init__(self, log):
        self.log = log

    def StreamStdio(self, request, context):
        while True:
            sd = grpc_stdio_pb2.StdioData(channel=1, data=self.log.read())
            time.sleep(0.2)
            yield sd


def serve():
    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    kv_pb2_grpc.add_KVServicer_to_server(KVServicer(), server)

    grpc_stdio_pb2_grpc.add_GRPCStdioServicer_to_server(StdioService(logger), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port('127.0.0.1:1234')
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
