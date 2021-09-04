// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vertex

import (
	"errors"
	"testing"

	"github.com/lasthyphen/paaro/snow/consensus/dijets"
)

var (
	errParse = errors.New("unexpectedly called Parse")

	_ Parser = &TestParser{}
)

type TestParser struct {
	T            *testing.T
	CantParseVtx bool
	ParseVtxF    func([]byte) (dijets.Vertex, error)
}

func (p *TestParser) Default(cant bool) { p.CantParseVtx = cant }

func (p *TestParser) ParseVtx(b []byte) (dijets.Vertex, error) {
	if p.ParseVtxF != nil {
		return p.ParseVtxF(b)
	}
	if p.CantParseVtx && p.T != nil {
		p.T.Fatal(errParse)
	}
	return nil, errParse
}
