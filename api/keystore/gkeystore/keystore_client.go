// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gkeystore

import (
	"context"

	"github.com/hashicorp/go-plugin"

	"github.com/lasthyphen/paaro/api/keystore"
	"github.com/lasthyphen/paaro/api/keystore/gkeystore/gkeystoreproto"
	"github.com/lasthyphen/paaro/database"
	"github.com/lasthyphen/paaro/database/encdb"
	"github.com/lasthyphen/paaro/database/rpcdb"
	"github.com/lasthyphen/paaro/database/rpcdb/rpcdbproto"
)

var _ keystore.BlockchainKeystore = &Client{}

// Client is a snow.Keystore that talks over RPC.
type Client struct {
	client gkeystoreproto.KeystoreClient
	broker *plugin.GRPCBroker
}

// NewClient returns a keystore instance connected to a remote keystore instance
func NewClient(client gkeystoreproto.KeystoreClient, broker *plugin.GRPCBroker) *Client {
	return &Client{
		client: client,
		broker: broker,
	}
}

func (c *Client) GetDatabase(username, password string) (*encdb.Database, error) {
	bcDB, err := c.GetRawDatabase(username, password)
	if err != nil {
		return nil, err
	}
	return encdb.New([]byte(password), bcDB)
}

func (c *Client) GetRawDatabase(username, password string) (database.Database, error) {
	resp, err := c.client.GetDatabase(context.Background(), &gkeystoreproto.GetDatabaseRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	dbConn, err := c.broker.Dial(resp.DbServer)
	if err != nil {
		return nil, err
	}

	dbClient := rpcdb.NewClient(rpcdbproto.NewDatabaseClient(dbConn))
	return dbClient, err
}
