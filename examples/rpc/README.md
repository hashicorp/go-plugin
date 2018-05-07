# RPC Plugin Example

This example shows the use of go-plugin

- Adds Restart Value to client config that will cause a plugin to be restarted if it exits
- Simplifies Bidirectional Communication by adding Serve(..) to client and Dispense(..) to server

To run build rpc/examples/extension.go and build/run host.go

Alternatively install go-task "go get -u -v github.com/go-task/task/cmd/task"
and run "task example" from this folder