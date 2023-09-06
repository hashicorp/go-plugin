# KV Example

This example builds a simple key/value store CLI where the mechanism
for storing and retrieving keys is pluggable. To build this example:

```sh
# This copies the Go WASM glue code
$ cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# This builds the main CLI
$ GOOS=js GOARCH=wasm go build -o kv.wasm

# This builds the plugin written in Go
$ GOOS=js GOARCH=wasm go build -o kv-go-grpc.wasm ./plugin-go-grpc

# This launches the HTTP server
# (install goexec: go install github.com/shurcooL/goexec)
goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'
```
