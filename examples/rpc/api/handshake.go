package api

import plugin "github.com/sampaioletti/go-plugin/examples/go-plugin"

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGGER",
	MagicCookieValue: "PLUGIT",
}
