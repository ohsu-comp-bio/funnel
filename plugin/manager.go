// Plugin manager used by the main application to load and invoke plugins.
package plugin

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/ohsu-comp-bio/funnel/auth"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// Manager loads and manages Auth plugins for this application.
//
// After creating a Manager value, call LoadPlugins with a directory path to
// discover and load plugins. At the end of the program call Close to kill and
// clean up all plugin processes.
type Manager struct {
	roleHooks     map[string]Authorizer
	contentsHooks []Authorizer

	pluginClients []*goplugin.Client
}

// LoadPlugins takes a directory path and assumes that all files within it
// are plugin binaries. It runs all these binaries in sub-processes,
// establishes RPC communication with the plugins, and registers them for
// the hooks they declare to support.
func (m *Manager) LoadPlugins(path string) error {
	m.contentsHooks = []Authorizer{}
	m.roleHooks = make(map[string]Authorizer)

	binaries, err := goplugin.Discover("*", path)
	if err != nil {
		return err
	}

	pluginMap := map[string]goplugin.Plugin{
		"authorize": &AuthorizePlugin{},
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Level: hclog.Warn,
	})

	for _, bpath := range binaries {
		client := goplugin.NewClient(&goplugin.ClientConfig{
			HandshakeConfig: Handshake,
			Plugins:         pluginMap,
			Cmd:             exec.Command(bpath),
			Logger:          logger,
		})
		m.pluginClients = append(m.pluginClients, client)

		rpcClient, err := client.Client()
		if err != nil {
			return err
		}

		raw, err := rpcClient.Dispense("authorize")
		if err != nil {
			return err
		}

		impl := raw.(Authorizer)

		// Query the plugin for its capabilities -- the hooks it supports.
		// Based on this information, register the plugin with the appropriate
		// role or contents hooks.
		capabilities := impl.Hooks()

		for _, cap := range capabilities {
			if cap == "contents" {
				m.contentsHooks = append(m.contentsHooks, impl)
			}
		}
	}

	return nil
}

func (m *Manager) Close() {
	for _, client := range m.pluginClients {
		client.Kill()
	}
}

// ApplyContentsHooks applies registered plugins to the given post contents,
// returning the transformed value. All registered plugins are applied in
// sequence to the value.
func (m *Manager) ApplyContentsHooks(authHeader http.Header, task *tes.Task) (auth.Auth, error) {
	for _, hook := range m.contentsHooks {

		resp, err := hook.Authorize(authHeader, task)

		if err != nil {
			return resp, err
		}

		return resp, nil
	}

	return auth.Auth{}, fmt.Errorf("No plugin found")
}
