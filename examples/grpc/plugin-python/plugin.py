# Copyright IBM Corp. 2016, 2025
# SPDX-License-Identifier: MPL-2.0

import sys
import threading
from concurrent import futures

import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
from grpc_health.v1.health import HealthServicer
from proto import (
    grpc_controller_pb2,
    grpc_controller_pb2_grpc,
    grpc_stdio_pb2_grpc,
    kv_pb2,
    kv_pb2_grpc,
)


class KVServicer(kv_pb2_grpc.KVServicer):
    """Implementation of KV service."""

    def Get(self, request, context):
        filename = "kv_" + request.key
        with open(filename, "r+b") as f:
            result = kv_pb2.GetResponse()
            result.value = f.read()
            return result

    def Put(self, request, context):
        filename = "kv_" + request.key
        value = "{0}\n\nWritten from plugin-python".format(
            request.value.decode("utf-8")
        )
        with open(filename, "w") as f:
            f.write(value)

        return kv_pb2.Empty()


class GRPCControllerServicer(grpc_controller_pb2_grpc.GRPCControllerServicer):
    """Implementation of the GRPCController service.

    go-plugin calls Shutdown on this service to ask the plugin to exit
    gracefully. Without it, the host has to fall back to killing the
    process after a timeout.
    """

    def __init__(self, shutdown_event):
        self._shutdown_event = shutdown_event

    def Shutdown(self, request, context):
        # Signal the main thread to stop the server, then acknowledge.
        self._shutdown_event.set()
        return grpc_controller_pb2.Empty()


class GRPCStdioServicer(grpc_stdio_pb2_grpc.GRPCStdioServicer):
    """Implementation of the GRPCStdio service.

    go-plugin connects to this service immediately to mirror the plugin's
    stdout/stderr on the host side. This example plugin doesn't emit any
    stdout/stderr of its own, so we simply hold the stream open until the
    host cancels it. Implementing the service avoids the host logging an
    "Unimplemented" / "Method not found" error.
    """

    def StreamStdio(self, request, context):
        # Block until the host cancels the stream (i.e. on plugin shutdown).
        # We yield no data; this is a generator so the return value is an
        # (empty) stream rather than a single message.
        cancelled = threading.Event()
        context.add_callback(cancelled.set)
        cancelled.wait()
        return
        yield  # pragma: no cover - makes this function a generator


def serve():
    # Event used to coordinate a graceful shutdown triggered by the host.
    shutdown_event = threading.Event()

    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value("SERVING"))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    kv_pb2_grpc.add_KVServicer_to_server(KVServicer(), server)
    grpc_controller_pb2_grpc.add_GRPCControllerServicer_to_server(
        GRPCControllerServicer(shutdown_event), server
    )
    grpc_stdio_pb2_grpc.add_GRPCStdioServicer_to_server(GRPCStdioServicer(), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port("127.0.0.1:1234")
    server.start()

    # Output information
    print("1|1|tcp|127.0.0.1:1234|grpc")
    sys.stdout.flush()

    try:
        # Wait until the host asks us to shut down (or we're interrupted).
        shutdown_event.wait()
    except KeyboardInterrupt:
        pass
    finally:
        server.stop(0)


if __name__ == "__main__":
    serve()
