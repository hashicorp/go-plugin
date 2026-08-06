
using System;
using System.IO;
using System.Threading.Tasks;
using Google.Protobuf;
using Grpc.Core;
using Proto;

namespace plugin_dotnet
{
    class Plugin : KV.KVBase
    {
        public const string ServiceHost = "localhost";
        public const int ServicePort = 1234;
        public const int AppProtoVersion = 1;

        public override async Task<Empty> Put(PutRequest request, ServerCallContext context)
        {
            var filename = $"kv_{request.Key}";
            await File.WriteAllTextAsync(filename,
                $"{request.Value.ToStringUtf8()}\n\nWritten from plugin-dotnet\n");
            
            return new Empty();
        }

        public override async Task<GetResponse> Get(GetRequest request, ServerCallContext context)
        {
            var filename = $"kv_{request.Key}";
            return new GetResponse
            {
                Value = ByteString.CopyFromUtf8(await File.ReadAllTextAsync(filename)),
            };
        }

        static async Task Main(string[] args)
        {
            // go-plugin semantics depend on the Health Check service from gRPC
            var health = HealthService.Get();
            health.SetStatus("plugin", HealthStatus.Serving);

            // Build a server to host the plugin over gRPC
            var server = new Server
            {
                Ports = { { ServiceHost, ServicePort, ServerCredentials.Insecure } },
                Services = {
                    { HealthService.BindService(health) },
                    { KV.BindService(new Plugin()) },
                },
            };

            server.Start();

            // Part of the go-plugin handshake:
            //  https://github.com/hashicorp/go-plugin/blob/master/docs/guide-plugin-write-non-go.md#4-output-handshake-information
            await Console.Out.WriteAsync($"1|1|tcp|{ServiceHost}:{ServicePort}|grpc\n");
            await Console.Out.FlushAsync();

            while (Console.Read() == -1)
                await Task.Delay(1000);
                
            await server.ShutdownAsync();
        }
    }
}
