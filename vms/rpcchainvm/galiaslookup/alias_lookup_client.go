// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package galiaslookup

import (
	"context"

	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow"
	"github.com/lasthyphen/paaro/vms/rpcchainvm/galiaslookup/galiaslookupproto"
)

var _ snow.AliasLookup = &Client{}

// Client implements alias lookups that talk over RPC.
type Client struct {
	client galiaslookupproto.AliasLookupClient
}

// NewClient returns an alias lookup instance connected to a remote alias lookup
// instance
func NewClient(client galiaslookupproto.AliasLookupClient) *Client {
	return &Client{client: client}
}

func (c *Client) Lookup(alias string) (ids.ID, error) {
	resp, err := c.client.Lookup(context.Background(), &galiaslookupproto.LookupRequest{
		Alias: alias,
	})
	if err != nil {
		return ids.ID{}, err
	}
	return ids.ToID(resp.Id)
}

func (c *Client) PrimaryAlias(id ids.ID) (string, error) {
	resp, err := c.client.PrimaryAlias(context.Background(), &galiaslookupproto.PrimaryAliasRequest{
		Id: id[:],
	})
	if err != nil {
		return "", err
	}
	return resp.Alias, nil
}
