Plugin Example
--------------

Copy the Go WASM glue file via:

    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

Compile the plugin itself via:

    GOOS=js GOARCH=wasm go build -o ./plugin/greeter ./plugin/greeter_impl.go

Compile this driver via:

    GOOS=js GOARCH=wasm go build -o basic .

You can then launch the plugin sample via:

    go run ./httpserve
