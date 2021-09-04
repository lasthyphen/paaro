// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vertex

import (
	"github.com/lasthyphen/paaro/snow/consensus/dijets"
	"github.com/lasthyphen/paaro/utils/hashing"
)

// Parser parses bytes into a vertex.
type Parser interface {
	// Parse a vertex from a slice of bytes
	ParseVtx(vertex []byte) (dijets.Vertex, error)
}

// Parse the provided vertex bytes into a stateless vertex
func Parse(vertex []byte) (StatelessVertex, error) {
	vtx := innerStatelessVertex{}
	version, err := c.Unmarshal(vertex, &vtx)
	vtx.Version = version
	return statelessVertex{
		innerStatelessVertex: vtx,
		id:                   hashing.ComputeHash256Array(vertex),
		bytes:                vertex,
	}, err
}
