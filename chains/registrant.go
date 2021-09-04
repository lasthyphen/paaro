// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chains

import (
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/snow/engine/common"
)

// Registrant can register the existence of a chain
type Registrant interface {
	// Called when the chain described by [ctx] and [engine] is created
	// This function is called before the chain starts processing messages
	// [engine] should be an dijets.Engine or snowman.Engine
	RegisterChain(name string, ctx *snow.Context, engine common.Engine)
}
