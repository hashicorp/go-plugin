package wasmrunner

// addrTranslator implements stateless identity functions, as the host and plugin
// run in the same context wrt WASM.
type addrTranslator struct{}

func (*addrTranslator) PluginToHost(pluginNet, pluginAddr string) (string, string, error) {
	return pluginNet, pluginAddr, nil
}

func (*addrTranslator) HostToPlugin(hostNet, hostAddr string) (string, string, error) {
	return hostNet, hostAddr, nil
}
