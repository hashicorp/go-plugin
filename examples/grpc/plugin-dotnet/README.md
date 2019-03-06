# KV Example

This example builds a plugin in C# for the [KV Example](https://github.com/hashicorp/go-plugin/tree/master/examples/grpc) in the [`go-plugin`](https://github.com/hashicorp/go-plugin) system over RPC.

To build and use this example, first follow the [directions](../README.md) for building the main
CLI for the example.

Next, build this C#-example after downloading and installing the *latest* [.NET Core SDK](https://dotnet.microsoft.com/download):

```pwsh

## From the project root
$ dotnet publish -c Release -o go-plugins

## Tell the KV CLI to use the C# plugin
$ export KV_PLUGIN="dotnet ./go-plugins/plugin-dotnet.dll"

## Write and Read
$ ../kv put hello world
$ ../kv get hello

```
