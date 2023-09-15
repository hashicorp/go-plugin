// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"sort"
	"strconv"
	"strings"

	hclog "github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
)

// CoreProtocolVersion is the ProtocolVersion of the plugin system itself.
// We will increment this whenever we change any protocol behavior. This
// will invalidate any prior plugins but will at least allow us to iterate
// on the core in a safe way. We will do our best to do this very
// infrequently.
const CoreProtocolVersion = 1

// HandshakeConfig is the configuration used by client and servers to
// handshake before starting a plugin connection. This is embedded by
// both ServeConfig and ClientConfig.
//
// In practice, the plugin host creates a HandshakeConfig that is exported
// and plugins then can easily consume it.
type HandshakeConfig struct {
	// ProtocolVersion is the version that clients must match on to
	// agree they can communicate. This should match the ProtocolVersion
	// set on ClientConfig when using a plugin.
	// This field is not required if VersionedPlugins are being used in the
	// Client or Server configurations.
	ProtocolVersion uint

	// MagicCookieKey and value are used as a very basic verification
	// that a plugin is intended to be launched. This is not a security
	// measure, just a UX feature. If the magic cookie doesn't match,
	// we show human-friendly output.
	MagicCookieKey   string
	MagicCookieValue string
}

// PluginSet is a set of plugins provided to be registered in the plugin
// server.
type PluginSet map[string]Plugin

// ServeConfig configures what sorts of plugins are served.
type ServeConfig struct {
	// HandshakeConfig is the configuration that must match clients.
	HandshakeConfig

	// TLSProvider is a function that returns a configured tls.Config.
	TLSProvider func() (*tls.Config, error)

	// Plugins are the plugins that are served.
	// The implied version of this PluginSet is the Handshake.ProtocolVersion.
	Plugins PluginSet

	// VersionedPlugins is a map of PluginSets for specific protocol versions.
	// These can be used to negotiate a compatible version between client and
	// server. If this is set, Handshake.ProtocolVersion is not required.
	VersionedPlugins map[int]PluginSet

	// GRPCServer should be non-nil to enable serving the plugins over
	// gRPC. This is a function to create the server when needed with the
	// given server options. The server options populated by go-plugin will
	// be for TLS if set. You may modify the input slice.
	//
	// Note that the grpc.Server will automatically be registered with
	// the gRPC health checking service. This is not optional since go-plugin
	// relies on this to implement Ping().
	GRPCServer func([]grpc.ServerOption) *grpc.Server

	// Logger is used to pass a logger into the server. If none is provided the
	// server will create a default logger.
	Logger hclog.Logger

	// Test, if non-nil, will put plugin serving into "test mode". This is
	// meant to be used as part of `go test` within a plugin's codebase to
	// launch the plugin in-process and output a ReattachConfig.
	//
	// This changes the behavior of the server in a number of ways to
	// accomodate the expectation of running in-process:
	//
	//   * The handshake cookie is not validated.
	//   * Stdout/stderr will receive plugin reads and writes
	//   * Connection information will not be sent to stdout
	//
	Test *ServeTestConfig
}

// ServeTestConfig configures plugin serving for test mode. See ServeConfig.Test.
type ServeTestConfig struct {
	// Context, if set, will force the plugin serving to end when cancelled.
	// This is only a test configuration because the non-test configuration
	// expects to take over the process and therefore end on an interrupt or
	// kill signal. For tests, we need to kill the plugin serving routinely
	// and this provides a way to do so.
	//
	// If you want to wait for the plugin process to close before moving on,
	// you can wait on CloseCh.
	Context context.Context

	// If this channel is non-nil, we will send the ReattachConfig via
	// this channel. This can be encoded (via JSON recommended) to the
	// plugin client to attach to this plugin.
	ReattachConfigCh chan<- *ReattachConfig

	// CloseCh, if non-nil, will be closed when serving exits. This can be
	// used along with Context to determine when the server is fully shut down.
	// If this is not set, you can still use Context on its own, but note there
	// may be a period of time between canceling the context and the plugin
	// server being shut down.
	CloseCh chan<- struct{}

	// SyncStdio, if true, will enable the client side "SyncStdout/Stderr"
	// functionality to work. This defaults to false because the implementation
	// of making this work within test environments is particularly messy
	// and SyncStdio functionality is fairly rare, so we default to the simple
	// scenario.
	SyncStdio bool
}

// protocolVersion determines the protocol version and plugin set to be used by
// the server. In the event that there is no suitable version, the last version
// in the config is returned leaving the client to report the incompatibility.
func protocolVersion(opts *ServeConfig) (int, Protocol, PluginSet) {
	protoVersion := int(opts.ProtocolVersion)
	pluginSet := opts.Plugins
	protoType := ProtocolNetRPC
	// Check if the client sent a list of acceptable versions
	var clientVersions []int
	if vs := os.Getenv("PLUGIN_PROTOCOL_VERSIONS"); vs != "" {
		for _, s := range strings.Split(vs, ",") {
			v, err := strconv.Atoi(s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "server sent invalid plugin version %q", s)
				continue
			}
			clientVersions = append(clientVersions, v)
		}
	}

	// We want to iterate in reverse order, to ensure we match the newest
	// compatible plugin version.
	sort.Sort(sort.Reverse(sort.IntSlice(clientVersions)))

	// set the old un-versioned fields as if they were versioned plugins
	if opts.VersionedPlugins == nil {
		opts.VersionedPlugins = make(map[int]PluginSet)
	}

	if pluginSet != nil {
		opts.VersionedPlugins[protoVersion] = pluginSet
	}

	// Sort the version to make sure we match the latest first
	var versions []int
	for v := range opts.VersionedPlugins {
		versions = append(versions, v)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(versions)))

	// See if we have multiple versions of Plugins to choose from
	for _, version := range versions {
		// Record each version, since we guarantee that this returns valid
		// values even if they are not a protocol match.
		protoVersion = version
		pluginSet = opts.VersionedPlugins[version]

		// If we have a configured gRPC server we should select a protocol
		if opts.GRPCServer != nil {
			// All plugins in a set must use the same transport, so check the first
			// for the protocol type
			for _, p := range pluginSet {
				switch p.(type) {
				case GRPCPlugin:
					protoType = ProtocolGRPC
				default:
					protoType = ProtocolNetRPC
				}
				break
			}
		}

		for _, clientVersion := range clientVersions {
			if clientVersion == protoVersion {
				return protoVersion, protoType, pluginSet
			}
		}
	}

	// Return the lowest version as the fallback.
	// Since we iterated over all the versions in reverse order above, these
	// values are from the lowest version number plugins (which may be from
	// a combination of the Handshake.ProtocolVersion and ServeConfig.Plugins
	// fields). This allows serving the oldest version of our plugins to a
	// legacy client that did not send a PLUGIN_PROTOCOL_VERSIONS list.
	return protoVersion, protoType, pluginSet
}

func serverListener(unixSocketCfg UnixSocketConfig) (net.Listener, error) {
	if runtime.GOOS == "windows" {
		return serverListener_tcp()
	}

	return serverListener_unix(unixSocketCfg)
}

func serverListener_tcp() (net.Listener, error) {
	envMinPort := os.Getenv("PLUGIN_MIN_PORT")
	envMaxPort := os.Getenv("PLUGIN_MAX_PORT")

	var minPort, maxPort int64
	var err error

	switch {
	case len(envMinPort) == 0:
		minPort = 0
	default:
		minPort, err = strconv.ParseInt(envMinPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get value from PLUGIN_MIN_PORT: %v", err)
		}
	}

	switch {
	case len(envMaxPort) == 0:
		maxPort = 0
	default:
		maxPort, err = strconv.ParseInt(envMaxPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get value from PLUGIN_MAX_PORT: %v", err)
		}
	}

	if minPort > maxPort {
		return nil, fmt.Errorf("PLUGIN_MIN_PORT value of %d is greater than PLUGIN_MAX_PORT value of %d", minPort, maxPort)
	}

	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
	}

	return nil, errors.New("Couldn't bind plugin TCP listener")
}

func serverListener_unix(unixSocketCfg UnixSocketConfig) (net.Listener, error) {
	tf, err := os.CreateTemp(unixSocketCfg.directory, "plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	// By default, unix sockets are only writable by the owner. Set up a custom
	// group owner and group write permissions if configured.
	if unixSocketCfg.Group != "" {
		err = setGroupWritable(path, unixSocketCfg.Group, 0o660)
		if err != nil {
			return nil, err
		}
	}

	// Wrap the listener in rmListener so that the Unix domain socket file
	// is removed on close.
	return &rmListener{
		Listener: l,
		Path:     path,
	}, nil
}

func setGroupWritable(path, groupString string, mode os.FileMode) error {
	groupID, err := strconv.Atoi(groupString)
	if err != nil {
		group, err := user.LookupGroup(groupString)
		if err != nil {
			return fmt.Errorf("failed to find gid from %q: %w", groupString, err)
		}
		groupID, err = strconv.Atoi(group.Gid)
		if err != nil {
			return fmt.Errorf("failed to parse %q group's gid as an integer: %w", groupString, err)
		}
	}

	err = os.Chown(path, -1, groupID)
	if err != nil {
		return err
	}

	err = os.Chmod(path, mode)
	if err != nil {
		return err
	}

	return nil
}

// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

func (l *rmListener) Close() error {
	// Close the listener itself
	if err := l.Listener.Close(); err != nil {
		return err
	}

	// Remove the file
	return os.Remove(l.Path)
}
