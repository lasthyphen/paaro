// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dijets

import (
	"github.com/lasthyphen/paaro/snow/choices"
	"github.com/lasthyphen/paaro/snow/consensus/snowstorm"
)

// Vertex is a collection of multiple transactions tied to other vertices
type Vertex interface {
	choices.Decidable

	// Returns the vertices this vertex depends on
	Parents() ([]Vertex, error)

	// Returns the height of this vertex. A vertex's height is defined by one
	// greater than the maximum height of the parents.
	Height() (uint64, error)

	// Returns the epoch this vertex was issued in.
	Epoch() (uint32, error)

	// Returns a series of state transitions to be performed on acceptance
	Txs() ([]snowstorm.Tx, error)

	// Returns the binary representation of this vertex
	Bytes() []byte
}
