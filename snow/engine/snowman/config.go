// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"github.com/djt-labs/paaro/snow/consensus/snowball"
	"github.com/djt-labs/paaro/snow/consensus/snowman"
	"github.com/djt-labs/paaro/snow/engine/snowman/bootstrap"
)

// Config wraps all the parameters needed for a snowman engine
type Config struct {
	bootstrap.Config

	Params    snowball.Parameters
	Consensus snowman.Consensus
}
