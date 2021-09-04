// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dvm

import (
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow"
)

// ID that this VM uses when labeled
var (
	ID = ids.ID{'d', 'v', 'm'}
)

type Factory struct {
	TxFee            uint64
	CreateAssetTxFee uint64
}

func (f *Factory) New(*snow.Context) (interface{}, error) {
	return &VM{Factory: *f}, nil
}
