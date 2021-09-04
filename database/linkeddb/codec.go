// (c) 2019-2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package linkeddb

import (
	"math"

	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/codec/linearcodec"
	"github.com/djt-labs/paaro/codec/reflectcodec"
)

const (
	codecVersion = 0
)

// c does serialization and deserialization
var (
	c codec.Manager
)

func init() {
	lc := linearcodec.New(reflectcodec.DefaultTagName, math.MaxUint32)
	c = codec.NewManager(math.MaxUint32)

	if err := c.RegisterCodec(codecVersion, lc); err != nil {
		panic(err)
	}
}
