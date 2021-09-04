// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dijets

import (
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow/consensus/dijets"
	"github.com/djt-labs/paaro/snow/engine/common"
)

// Engine describes the events that can occur on a consensus instance
type Engine interface {
	common.Engine

	// Initialize this engine.
	Initialize(Config) error

	// GetVtx returns a vertex by its ID.
	// Returns an error if unknown.
	GetVtx(vtxID ids.ID) (dijets.Vertex, error)
}
