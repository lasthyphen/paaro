// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dijets

import (
	"github.com/lasthyphen/paaro/snow/consensus/dijets"
	"github.com/lasthyphen/paaro/snow/engine/dijets/bootstrap"
)

// Config wraps all the parameters needed for an dijets engine
type Config struct {
	bootstrap.Config

	Params    dijets.Parameters
	Consensus dijets.Consensus
}
