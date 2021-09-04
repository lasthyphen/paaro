// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package plugin

import (
	"context"

	appproto "github.com/djt-labs/paaro/app/plugin/proto"
	"github.com/djt-labs/paaro/app/process"
	"google.golang.org/grpc"

	"github.com/hashicorp/go-plugin"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "NODE_PROCESS_PLUGIN",
	MagicCookieValue: "dynamic",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"nodeProcess": &AppPlugin{},
}

// AppPlugin is can be served/consumed with the hashicorp plugin library.
// Plugin implements plugin.GRPCPlugin
type AppPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	app *process.App
}

func New(app *process.App) *AppPlugin {
	return &AppPlugin{
		app: app,
	}
}

// GRPCServer registers a new GRPC server.
func (p *AppPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	appproto.RegisterNodeServer(s, NewServer(p.app))
	return nil
}

// GRPCClient returns a new GRPC client
func (p *AppPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return NewClient(appproto.NewNodeClient(c)), nil
}
