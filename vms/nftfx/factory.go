package nftfx

import (
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
)

// ID that this Fx uses when labeled
var (
	ID = ids.ID{'n', 'f', 't', 'f', 'x'}
)

type Factory struct{}

func (f *Factory) New(*snow.Context) (interface{}, error) { return &Fx{}, nil }
