package nftfx

import (
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow"
)

// ID that this Fx uses when labeled
var (
	ID = ids.ID{'n', 'f', 't', 'f', 'x'}
)

type Factory struct{}

func (f *Factory) New(*snow.Context) (interface{}, error) { return &Fx{}, nil }
