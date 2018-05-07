package api

import plugin "gitlab.com/indis/libs/third_party/go-plugin"

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGGER",
	MagicCookieValue: "PLUGIT",
}
