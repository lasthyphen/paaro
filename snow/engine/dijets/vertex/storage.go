// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vertex

import (
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow/consensus/dijets"
)

// Storage defines the persistent storage that is required by the consensus
// engine.
type Storage interface {
	// Get a vertex by its hash from storage.
	GetVtx(vtxID ids.ID) (dijets.Vertex, error)

	// Edge returns a list of accepted vertex IDs with no accepted children.
	Edge() (vtxIDs []ids.ID)
}
