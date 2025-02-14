// Defines the Handshake type for this plugin.
package plugin

import goplugin "github.com/hashicorp/go-plugin"

// Handshake is a common handshake that is shared by plugin and host, to ensure
// their versions match.
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "AUTH_PLUGIN",
	MagicCookieValue: "exampleCookieValue",
}
